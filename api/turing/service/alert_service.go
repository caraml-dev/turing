package service

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"text/template"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/config"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
)

type Metric string

type AlertService interface {
	Save(alert models.Alert, authorEmail string, dashboardURL string) (*models.Alert, error)
	List(service string) ([]*models.Alert, error)
	FindByID(id models.ID) (*models.Alert, error)
	Update(alert models.Alert, authorEmail string, dashboardURL string) error
	Delete(alert models.Alert, authorEmail string, dashboardURL string) error
	// GetDashboardURL returns the dashboard URL for the router alert.
	//
	// If routerVersion is nil, the dashboard URL should return the dashboard showing metrics
	// for the router across all versions. Else, the dashboard should show metrics for a specific
	// router version.
	GetDashboardURL(
		alert *models.Alert,
		project *mlp.Project,
		environment *merlin.Environment,
		router *models.Router,
		routerVersion *models.RouterVersion) (string, error)
}

type gitlabOpsAlertService struct {
	db     *gorm.DB       // database client
	gitlab *gitlab.Client // GitLab client
	// dashboardURLTemplate is a template for grafana dashboard URL that shows router metric.
	// The template will be executed with dashboardURLValue.
	dashboardURLTemplate template.Template
	config               config.AlertConfig
}

// DashboardURLValue dashboardURLValue will be passed in as argument to execute dashboardURLTemplate.
type DashboardURLValue struct {
	Environment string // environment name where the router is deployed
	Cluster     string // Kubernetes cluster name where the router name is deployed
	Project     string // MLP project name where the router is deployed
	Router      string // router name for the alert
	Version     string // router version number
}

// NewGitlabOpsAlertService Creates a new AlertService that can be used with GitOps based on GitLab. It is assumed
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
func NewGitlabOpsAlertService(db *gorm.DB, config config.AlertConfig) (AlertService, error) {
	if config.GitLab == nil {
		return nil, errors.New("missing GitLab AlertConfig")
	}

	gitlabClient, err := gitlab.NewClient(config.GitLab.Token, gitlab.WithBaseURL(config.GitLab.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client required for alerting: %v", err)
	}

	tmpl, err := template.New("dashboardURLTemplate").Parse(config.DashboardURLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dashboardUrlTemplate: %v", err)
	}

	return &gitlabOpsAlertService{
		db:                   db,
		gitlab:               gitlabClient,
		dashboardURLTemplate: *tmpl,
		config:               config,
	}, nil
}

// Save will persist the alert in the database and the configured GitLab alert repository.
// The Git file creation will be committed by "authorEmail". Save will fail if either GitLab
// or the database is down.
func (service *gitlabOpsAlertService) Save(
	alert models.Alert, authorEmail string, dashboardURL string) (*models.Alert, error) {
	if err := alert.Validate(); err != nil {
		return nil, fmt.Errorf("alert is invalid: %s", err)
	}
	if err := service.createInGitLab(alert, authorEmail, dashboardURL); err != nil {
		return nil, fmt.Errorf("failed to create alert in GitLab: %s", err)
	}
	if err := service.db.Create(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the file creation in GitLab
		if err := service.deleteInGitLab(alert, authorEmail); err != nil {
			log.Errorf("failed to revert alert creation in GitLab, "+
				"file '%s' in GitLab should be deleted manually: %s",
				getGitLabFilePath(service.config.GitLab.PathPrefix, alert), err)
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

// FindByID Find an alert by its ID. An error will be returned if no alert is found.
func (service *gitlabOpsAlertService) FindByID(id models.ID) (*models.Alert, error) {
	var alert models.Alert
	if err := service.db.Where("id = ?", id).First(&alert).Error; err != nil {
		return nil, fmt.Errorf("failed to find alert with id '%d' in the database: %s", id, err)
	}
	return &alert, nil
}

// Update an alert with the new alert object. The new alert must contain an existing alert ID
// and have all the required fields populated.
func (service *gitlabOpsAlertService) Update(alert models.Alert, authorEmail string, dashboardURL string) error {
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("alert is invalid: %s", err)
	}
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.updateInGitLab(alert, authorEmail, dashboardURL); err != nil {
		return err
	}
	if err := service.db.Save(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the update in GitLab
		if err := service.updateInGitLab(*oldAlert, authorEmail, dashboardURL); err != nil {
			log.Errorf("failed to revert alert update in GitLab, "+
				"file '%s' in GitLab should be reverted to previous state manually: %s",
				getGitLabFilePath(service.config.GitLab.PathPrefix, alert), err)
		}
		return fmt.Errorf("failed to update alert in the database: %s", err)
	}
	return nil
}

// Delete an alert by the ID of the alert object in the argument.
func (service *gitlabOpsAlertService) Delete(alert models.Alert, authorEmail string, dashboardURL string) error {
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.deleteInGitLab(alert, authorEmail); err != nil {
		return err
	}
	if err := service.db.Delete(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the deletion in GitLab
		if err := service.createInGitLab(*oldAlert, authorEmail, dashboardURL); err != nil {
			log.Errorf("failed to revert alert deletion in GitLab, "+
				"file '%s' in GitLab should be re-created manually: %s",
				getGitLabFilePath(service.config.GitLab.PathPrefix, alert), err)
		}
		return fmt.Errorf("failed to delete alert in the database: %s", err)
	}
	return nil
}

func (service *gitlabOpsAlertService) GetDashboardURL(
	alert *models.Alert,
	project *mlp.Project,
	environment *merlin.Environment,
	router *models.Router,
	routerVersion *models.RouterVersion) (string, error) {
	var version string
	if routerVersion != nil {
		version = fmt.Sprintf("%d", routerVersion.Version)
	} else {
		version = "$__all"
	}

	value := DashboardURLValue{
		Environment: alert.Environment,
		Cluster:     environment.Cluster,
		Project:     project.Name,
		Router:      router.Name,
		Version:     version,
	}

	var buf bytes.Buffer
	err := service.dashboardURLTemplate.Execute(&buf, value)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (service *gitlabOpsAlertService) createInGitLab(
	alert models.Alert,
	authorEmail string,
	dashboardURL string) error {
	alertGroups, err := yaml.Marshal(struct {
		Groups []interface{} `yaml:"groups"`
	}{
		Groups: []interface{}{alert.Group(service.config.PlaybookURL, dashboardURL)},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal AlertGroup into yaml: %s", err)
	}
	fileContent := string(alertGroups)
	filePath := getGitLabFilePath(service.config.GitLab.PathPrefix, alert)

	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, _, err = service.gitlab.RepositoryFiles.CreateFile(
		service.config.GitLab.ProjectID,
		filePath,
		&gitlab.CreateFileOptions{
			Branch:        &service.config.GitLab.Branch,
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

func (service *gitlabOpsAlertService) updateInGitLab(
	alert models.Alert,
	authorEmail string,
	dashboardURL string) error {
	alertGroups, err := yaml.Marshal(struct {
		Groups []interface{} `yaml:"groups"`
	}{
		Groups: []interface{}{alert.Group(service.config.PlaybookURL, dashboardURL)},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal AlertGroup into yaml: %s", err)
	}
	fileContent := string(alertGroups)
	filePath := getGitLabFilePath(service.config.GitLab.PathPrefix, alert)

	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, _, err = service.gitlab.RepositoryFiles.UpdateFile(
		service.config.GitLab.ProjectID,
		filePath,
		&gitlab.UpdateFileOptions{
			Branch:        &service.config.GitLab.Branch,
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
	filePath := getGitLabFilePath(service.config.GitLab.PathPrefix, alert)
	commitMessage := fmt.Sprintf("Autogenerated alert for service: %s, team: %s, environment: %s",
		alert.Service, alert.Team, alert.Environment)

	_, err := service.gitlab.RepositoryFiles.DeleteFile(
		service.config.GitLab.ProjectID,
		filePath,
		&gitlab.DeleteFileOptions{
			Branch:        &service.config.GitLab.Branch,
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
