package resultlog

import (
	"github.com/fluent/fluent-logger-golang/fluent"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"

	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
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
	bqLogger     BigQueryLogger
	fluentLogger fluentdClient
}

// NewFluentdLogger creates a new FluentdLogger
func NewFluentdLogger(
	cfg *config.FluentdConfig,
	bqLogger BigQueryLogger,
) (*FluentdLogger, error) {
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
		bqLogger:     bqLogger,
		fluentLogger: fClient,
	}, nil
}

// write satisfies the TuringResultLogger interface. Fluentd logs are synced to a BigQuery
// output destination and hence, calling write() uses the BigQueryLogger to generate a
// loggable record, of the required schema, that is posted to a Fluentd server.
func (l *FluentdLogger) write(turLogEntry *turing.TuringResultLogMessage) error {
	// Measure time taken to post the log to fluentd
	var err error
	defer metrics.Glob().MeasureDurationMs(
		instrumentation.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "fluentd_post"
			},
			"traffic_rule": func() string { return "" },
		},
	)()
	log.Glob().Debugw("Sending log to fluentd", "entry", turLogEntry)
	err = l.fluentLogger.Post(l.tag, l.bqLogger.getLogData(turLogEntry))
	return err
}
