package worker

import (
	"time"
)

// JobRunner is an interface that exposes the batch processes.
type JobRunner interface {
	Run()
	GetInterval() time.Duration
}

// Start will run all runners asynchronously.
func Start(runners []JobRunner) {
	for _, runner := range runners {
		go func(runner JobRunner) {
			for {
				runner.Run()
				time.Sleep(runner.GetInterval())
			}
		}(runner)
	}
}
