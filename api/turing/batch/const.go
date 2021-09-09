package batch

const (
	// JobConfigFileName is the name of the defined container job config, e.g. ensembler config
	JobConfigFileName = "jobConfig.yaml"
	// JobConfigMount is where the job spec if mounted
	JobConfigMount = "/mnt/job-spec/"
)

const (
	// DatasetTypeBQ is the BQ source dataset type
	DatasetTypeBQ = "BQ"
	// SinkTypeBQ is the BQ sink dataset type
	SinkTypeBQ = "BQ"

	// Batch Ensembling Pod component types

	// ImageBuilderPodType is the image builder pod type
	ImageBuilderPodType = "image_builder"
	// DriverPodType is the spark driver pod type
	DriverPodType = "driver"
	// ExecutorPodType is the spark executor pod type
	ExecutorPodType = "executor"
)
