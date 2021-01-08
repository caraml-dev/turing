package resultlog

import (
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/metrics"
)

// fluentdClient minimally defines the functionality used by the FluentdLogger for sending
// the result log to a fluentd server (useful for mocking in tests).
type fluentdClient interface {
	Post(string, interface{}) error
}

// FluentdLogger generates instances of FluentdLog for posting results to a
// fluentd backend.
type FluentdLogger struct {
	tag          string
	fluentLogger fluentdClient
}

// newFluentdLogger creates a new FluentdLogger
func newFluentdLogger(cfg *config.FluentdConfig) (*FluentdLogger, error) {
	fClient, err := fluent.New(fluent.Config{
		FluentHost: cfg.Host,
		FluentPort: cfg.Port,
	})
	if err != nil {
		return nil, err
	}
	// Check if the tag is set
	if cfg.Tag == "" {
		return nil, errors.Newf(errors.BadConfig, "Fluentd Tag must be configured")
	}
	// Create FluentdLogger
	return &FluentdLogger{
		tag:          cfg.Tag,
		fluentLogger: fClient,
	}, nil
}

// write satisfies the TuringResultLogger interface. Fluentd logs are synced to a BigQuery
// output destination.
func (l *FluentdLogger) write(turLogEntry *TuringResultLogEntry) error {
	// Measure time taken to post the log to fluentd
	var err error
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "fluentd_post"
			},
		},
	)()

	// Convert to a generic map of key-value pairs for logging
	var kvPairs map[string]interface{}
	kvPairs, err = turLogEntry.Value()
	if err != nil {
		return err
	}

	err = l.fluentLogger.Post(l.tag, kvPairs)
	return err
}
