package batchrunner

import (
	"time"
)

// BatchJobRunner is an interface that exposes the batch processes.
type BatchJobRunner interface {
	Run()
}

// RunBatchRunners will run all runners asynchronously.
func RunBatchRunners(interval time.Duration, runners []BatchJobRunner) {
	for {
		for _, runner := range runners {
			go runner.Run()
		}
		time.Sleep(interval)
	}
}
