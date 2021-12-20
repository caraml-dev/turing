package manager

type GetExperimentRunnerConfigRequest struct {
	// Data is either json.RawMessage or manager.TuringExperimentConfig
	Data interface{}
}
