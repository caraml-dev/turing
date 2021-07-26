package batch

const (
	// JobConfigFileName is the name of the defined container job config, e.g. ensembler config
	JobConfigFileName = "jobConfig.yaml"
	// JobConfigMount is where the job spec if mounted
	JobConfigMount = "/mnt/job-spec/"
)

const (
	DatasetTypeBQ = "BQ"

	SinkTypeBQ = "BQ"
)
