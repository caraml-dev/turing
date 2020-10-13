package service

import (
	"fmt"
	"path"

	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

type Metric string

type AlertService interface {
	Save(alert models.Alert, authorEmail string) (*models.Alert, error)
	List(service string) ([]*models.Alert, error)
	FindByID(id uint) (*models.Alert, error)
	Update(alert models.Alert, authorEmail string) error
	Delete(alert models.Alert, authorEmail string) error
}

type gitlabOpsAlertService struct {
	db               *gorm.DB       // database client
	gitlab           *gitlab.Client // GitLab client
	gitlabProjectID  string         // GitLab project ID where the alert will be committed to
	gitlabBranch     string         // GitLab branch where the alert will be committed to. The branch must already exist
	gitlabPathPrefix string         // The alert file will be created under the gitlabPathPrefix folder
}

// Create a new AlertService that can be used with GitOps based on GitLab. It is assumed
// that the continuous integration (CI) jobs configured in GitLab can process the committed alert
// files to register the alerts to the corresponding external alert manager. This CI process
// is out of scope of Turing.
//
// For every alert object created, a yaml file will be created at the following location:
// <gitlab_alert_repo_root>/<gitlabPathPrefix>/<environment>/<team>/<service>/<metric>.yaml
//
// Every alert created will be persisted in the database and the configured
// alert repository in GitLab. In most cases, the alerts persisted in the database will be in sync
// with the alert files in GitLab (as long as the Git files are not manually modified i.e.
// the alert files are only updated by calling this service).
//
func NewGitlabOpsAlertService(db *gorm.DB, gitlab *gitlab.Client,
	gitlabProjectID string, gitlabBranch string, gitlabPathPrefix string) AlertService {
	return &gitlabOpsAlertService{
		db:               db,
		gitlab:           gitlab,
		gitlabBranch:     gitlabBranch,
		gitlabProjectID:  gitlabProjectID,
		gitlabPathPrefix: gitlabPathPrefix,
	}
}

// Save will persist the alert in the database and the configured GitLab alert repository.
// The Git file creation will be committed by "authorEmail". Save will fail if either GitLab
// or the database is down.
func (service *gitlabOpsAlertService) Save(alert models.Alert, authorEmail string) (*models.Alert, error) {
	if err := alert.Validate(); err != nil {
		return nil, fmt.Errorf("alert is invalid: %s", err)
	}
	if err := service.createInGitLab(alert, authorEmail); err != nil {
		return nil, fmt.Errorf("failed to create alert in GitLab: %s", err)
	}
	if err := service.db.Create(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the file creation in GitLab
		if err := service.deleteInGitLab(alert, authorEmail); err != nil {
			log.Errorf("failed to revert alert creation in GitLab, "+
				"file '%s' in GitLab should be deleted manually: %s",
				getGitLabFilePath(service.gitlabPathPrefix, alert), err)
		}
		return nil, fmt.Errorf("failed to create alert in the database: %s", err)
	}
	return service.FindByID(alert.ID)
}

// List the alerts by the service name. No error is returned for empty results.
func (service *gitlabOpsAlertService) List(svc string) ([]*models.Alert, error) {
	alerts := make([]*models.Alert, 0)
	if err := service.db.Where("service = ?", svc).Find(&alerts).Error; err != nil {
		return alerts, fmt.Errorf("failed to list alerts in the database: %s", err)
	}
	return alerts, nil
}

// Find an alert by its ID. An error will be returned if no alert is found.
func (service *gitlabOpsAlertService) FindByID(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := service.db.Where("id = ?", id).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to find alert with id '%d' in the database: %s", id, err)
	}
	return &alert, nil
}

// Update an alert with the new alert object. The new alert must contain an existing alert ID
// and have all the required fields populated.
func (service *gitlabOpsAlertService) Update(alert models.Alert, authorEmail string) error {
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("alert is invalid: %s", err)
	}
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.updateInGitLab(alert, authorEmail); err != nil {
		return err
	}
	if err := service.db.Save(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the update in GitLab
		if err := service.updateInGitLab(*oldAlert, authorEmail); err != nil {
			log.Errorf("failed to revert alert update in GitLab, "+
				"file '%s' in GitLab should be reverted to previous state manually: %s",
				getGitLabFilePath(service.gitlabPathPrefix, alert), err)
		}
		return fmt.Errorf("failed to update alert in the database: %s", err)
	}
	return nil
}

// Delete an alert by the ID of the alert object in the argument.
func (service *gitlabOpsAlertService) Delete(alert models.Alert, authorEmail string) error {
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.deleteInGitLab(alert, authorEmail); err != nil {
		return err
	}
	if err := service.db.Delete(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the deletion in GitLab
		if err := service.createInGitLab(*oldAlert, authorEmail); err != nil {
			log.Errorf("failed to revert alert deletion in GitLab, "+
				"file '%s' in GitLab should be re-created manually: %s",
				getGitLabFilePath(service.gitlabPathPrefix, alert), err)
		}
		return fmt.Errorf("failed to delete alert in the database: %s", err)
	}
	return nil
}

func (service *gitlabOpsAlertService) createInGitLab(alert models.Alert, authorEmail string) error {
	alertGroup, err := yaml.Marshal(alert.Group())
	if err != nil {
		return fmt.Errorf("failed to marshal AlertGroup into yaml: %s", err)
	}
	fileContent := string(alertGroup)
	filePath := getGitLabFilePath(service.gitlabPathPrefix, alert)

	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, _, err = service.gitlab.RepositoryFiles.CreateFile(
		service.gitlabProjectID,
		filePath,
		&gitlab.CreateFileOptions{
			Branch:        &service.gitlabBranch,
			AuthorEmail:   &authorEmail,
			Content:       &fileContent,
			CommitMessage: &commitMessage,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (service *gitlabOpsAlertService) updateInGitLab(alert models.Alert, authorEmail string) error {
	alertGroup, err := yaml.Marshal(alert.Group())
	if err != nil {
		return fmt.Errorf("failed to marshal AlertGroup into yaml: %s", err)
	}
	fileContent := string(alertGroup)
	filePath := getGitLabFilePath(service.gitlabPathPrefix, alert)

	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, _, err = service.gitlab.RepositoryFiles.UpdateFile(
		service.gitlabProjectID,
		filePath,
		&gitlab.UpdateFileOptions{
			Branch:        &service.gitlabBranch,
			AuthorEmail:   &authorEmail,
			Content:       &fileContent,
			CommitMessage: &commitMessage,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (service *gitlabOpsAlertService) deleteInGitLab(alert models.Alert, authorEmail string) error {
	filePath := getGitLabFilePath(service.gitlabPathPrefix, alert)
	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, err := service.gitlab.RepositoryFiles.DeleteFile(
		service.gitlabProjectID,
		filePath,
		&gitlab.DeleteFileOptions{
			Branch:        &service.gitlabBranch,
			AuthorEmail:   &authorEmail,
			CommitMessage: &commitMessage,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func getGitLabFilePath(prefix string, alert models.Alert) string {
	return path.Join(
		prefix,
		alert.Environment,
		alert.Team,
		alert.Service,
		fmt.Sprintf("%s.yaml", alert.Metric),
	)
}
