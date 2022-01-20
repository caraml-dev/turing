package service

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/gojek/turing/api/turing/models"
)

type RouterMonitoringService interface {
	GenerateMonitoringURL(
		projectID models.ID,
		environmentName string,
		routerName string,
		routerVersion *uint,
	) (string, error)
}

type routerMonitoringService struct {
	mlpService            MLPService
	monitoringURLTemplate *template.Template
}

type monitoringURLValues struct {
	ClusterName string
	ProjectName string
	RouterName  string
	Version     string
}

func NewRouterMonitoringService(
	mlpService MLPService,
	monitoringURLTemplate *template.Template) RouterMonitoringService {
	return &routerMonitoringService{
		mlpService:            mlpService,
		monitoringURLTemplate: monitoringURLTemplate,
	}
}

// GenerateMonitoringURL generates the monitoring url based on the router version.
func (service *routerMonitoringService) GenerateMonitoringURL(
	projectID models.ID,
	environmentName string,
	routerName string,
	routerVersion *uint,
) (string, error) {
	if service.monitoringURLTemplate == nil {
		return "", nil
	}

	project, err := service.mlpService.GetProject(projectID)
	if err != nil {
		return "", err
	}

	env, err := service.mlpService.GetEnvironment(environmentName)
	if err != nil {
		return "", err
	}

	var versionString string
	if routerVersion == nil {
		versionString = grafanaAllVariable
	} else {
		versionString = fmt.Sprintf("%d", *routerVersion)
	}

	values := monitoringURLValues{
		ClusterName: env.Cluster,
		ProjectName: project.Name,
		RouterName:  routerName,
		Version:     versionString,
	}

	var b bytes.Buffer
	err = service.monitoringURLTemplate.Execute(&b, values)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
