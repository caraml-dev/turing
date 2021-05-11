package batch

// JobRunner is an interface that exposes the batch processes.
type JobRunner interface {
	Run()
}

// Controller is an interface that exposes the batch kubernetes controller.
type Controller interface {
	Create() error
}
