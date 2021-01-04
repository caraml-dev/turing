package service

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"text/template"

	"github.com/gojek/turing/api/turing/config"

	"github.com/gojek/turing/api/turing/config"

	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
)

type Metric string

type AlertService interface {
	Save(alert models.Alert, router models.Router, authorEmail string) (*models.Alert, error)
	List(service string) ([]*models.Alert, error)
	FindByID(id uint) (*models.Alert, error)
	Update(alert models.Alert, router models.Router, authorEmail string) error
	Delete(alert models.Alert, router models.Router, authorEmail string) error
	// GetDashboardURL returns the Grafana dashboard URL with router metrics that can be used to debug alerts.
	// If routerVersion is nil, the dashboard should show router metrics for all versions, else
	// the dashboard should show metrics for a particular router version.
	GetDashboardURL(router *models.Router, routerVersion *models.RouterVersion) (string, error)
}

type gitlabOpsAlertService struct {
	db     *gorm.DB       // database client
	gitlab *gitlab.Client // GitLab client
	// mlpService is used to get cluster name and project name for the router that corresponds to the alert
	mlpService MLPService
	// dashboardURLTemplate is a template for grafana dashboard URL that shows router metric.
	// The template will be executed with dashboardURLValue.
	dashboardURLTemplate template.Template
	config               config.AlertConfig
}

// dashboardURLValue will be passed in as argument to execute dashboardURLTemplate.
type dashboardURLValue struct {
	Environment string // environment name where the router is deployed
	Cluster     string // Kubernetes cluster name where the router name is deployed
	Project     string // MLP project name where the router is deployed
	Router      string // router name for the alert
	Revision    string // router revision number
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
func NewGitlabOpsAlertService(db *gorm.DB, mlpService MLPService, config config.AlertConfig) (AlertService, error) {
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
		mlpService:           mlpService,
		dashboardURLTemplate: *tmpl,
		config:               config,
	}, nil
}

// Save will persist the alert in the database and the configured GitLab alert repository.
// The Git file creation will be committed by "authorEmail". Save will fail if either GitLab
// or the database is down.
func (service *gitlabOpsAlertService) Save(
	alert models.Alert,
	router models.Router,
	authorEmail string) (*models.Alert, error) {
	if err := alert.Validate(); err != nil {
		return nil, fmt.Errorf("alert is invalid: %s", err)
	}
	if err := service.createInGitLab(alert, router, authorEmail); err != nil {
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
func (service *gitlabOpsAlertService) Update(alert models.Alert, router models.Router, authorEmail string) error {
	if err := alert.Validate(); err != nil {
		return fmt.Errorf("alert is invalid: %s", err)
	}
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.updateInGitLab(alert, router, authorEmail); err != nil {
		return err
	}
	if err := service.db.Save(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the update in GitLab
		if err := service.updateInGitLab(*oldAlert, router, authorEmail); err != nil {
			log.Errorf("failed to revert alert update in GitLab, "+
				"file '%s' in GitLab should be reverted to previous state manually: %s",
				getGitLabFilePath(service.config.GitLab.PathPrefix, alert), err)
		}
		return fmt.Errorf("failed to update alert in the database: %s", err)
	}
	return nil
}

// Delete an alert by the ID of the alert object in the argument.
func (service *gitlabOpsAlertService) Delete(alert models.Alert, router models.Router, authorEmail string) error {
	oldAlert, err := service.FindByID(alert.ID)
	if err != nil {
		return err
	}
	if err := service.deleteInGitLab(alert, authorEmail); err != nil {
		return err
	}
	if err := service.db.Delete(&alert).Error; err != nil {
		// If failed to save to DB, we should revert the deletion in GitLab
		if err := service.createInGitLab(*oldAlert, router, authorEmail); err != nil {
			log.Errorf("failed to revert alert deletion in GitLab, "+
				"file '%s' in GitLab should be re-created manually: %s",
				getGitLabFilePath(service.config.GitLab.PathPrefix, alert), err)
		}
		return fmt.Errorf("failed to delete alert in the database: %s", err)
	}
	return nil
}

func (service *gitlabOpsAlertService) GetDashboardURL(
	router *models.Router,
	routerVersion *models.RouterVersion) (string, error) {
	environment, err := service.mlpService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return "", err
	}

	project, err := service.mlpService.GetProject(router.ProjectID)
	if err != nil {
		return "", err
	}

	var revision string
	if routerVersion != nil {
		revision = fmt.Sprintf("%d", routerVersion.Version)
	} else {
		revision = "$__all"
	}

	value := dashboardURLValue{
		Environment: router.EnvironmentName,
		Cluster:     environment.Cluster,
		Project:     project.Name,
		Router:      router.Name,
		Revision:    revision,
	}

	var buf bytes.Buffer
	err = service.dashboardURLTemplate.Execute(&buf, value)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (service *gitlabOpsAlertService) createInGitLab(
	alert models.Alert,
	router models.Router,
	authorEmail string) error {
	dashboardURL, err := service.GetDashboardURL(&router, nil)
	if err != nil {
		return err
	}

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
	router models.Router,
	authorEmail string) error {
	dashboardURL, err := service.GetDashboardURL(&router, nil)
	if err != nil {
		return err
	}

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
