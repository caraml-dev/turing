package runner

import (
	"context"
	"net/http"
)

func NewInterceptRunner(
	name string,
	runner ExperimentRunner,
	interceptors ...Interceptor,
) ExperimentRunner {
	return &interceptRunner{
		ExperimentRunner: runner,
		name:             name,
		interceptors:     interceptors,
	}
}

type interceptRunner struct {
	ExperimentRunner

	name         string
	interceptors []Interceptor
}

func (r *interceptRunner) GetTreatmentForRequest(
	header http.Header,
	payload []byte,
	options GetTreatmentOptions,
) (*Treatment, error) {
	ctx := context.WithValue(context.Background(), ExperimentEngineKey, r.name)

	// Call BeforeDispatch on the interceptors, run the experiment and then AfterCompletion
	for _, interceptor := range r.interceptors {
		ctx = interceptor.BeforeDispatch(ctx)
	}

	treatment, err := r.ExperimentRunner.GetTreatmentForRequest(header, payload, options)

	for _, interceptor := range r.interceptors {
		interceptor.AfterCompletion(ctx, err)
	}

	return treatment, err
}
