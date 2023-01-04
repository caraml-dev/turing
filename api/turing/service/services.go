package service

type Services struct {
	AlertService            AlertService
	CryptoService           CryptoService
	EnsemblersService       EnsemblersService
	EnsemblingJobService    EnsemblingJobService
	EventService            EventService
	ExperimentsService      ExperimentsService
	MLPService              MLPService
	PodLogService           PodLogService
	RoutersService          RoutersService
	RouterVersionsService   RouterVersionsService
	RouterMonitoringService RouterMonitoringService
	RouterDeploymentService RouterDeploymentService
}
