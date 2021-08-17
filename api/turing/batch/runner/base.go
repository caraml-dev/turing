package batchrunner

import (
	"time"
)

// BatchJobRunner is an interface that exposes the batch processes.
type BatchJobRunner interface {
	Run()
	GetInterval() time.Duration
}

// RunBatchRunners will run all runners asynchronously.
func RunBatchRunners(runners []BatchJobRunner) {
	for _, runner := range runners {
		go func(runner BatchJobRunner) {
			for {
				runner.Run()
				time.Sleep(runner.GetInterval())
			}
		}(runner)
	}
}
