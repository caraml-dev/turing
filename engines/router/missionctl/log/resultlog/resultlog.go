package resultlog

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/log"
)

// Init the global logger to Nop Logger, calling InitTuringResultLogger will reset this.
var globalLogger TuringResultLogger = newNopLogger()

// ResultLogKeys defines the individual components for which the result log must be created
var ResultLogKeys = struct {
	Experiment string
	Enricher   string
	Router     string
	Ensembler  string
}{
	Experiment: "experiment",
	Enricher:   "enricher",
	Router:     "router",
	Ensembler:  "ensembler",
}

type requestLogEntry struct {
	Header *http.Header    `json:"header"`
	Body   json.RawMessage `json:"body"`
}

type responseLogEntry struct {
	Response json.RawMessage `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// TuringResultLogEntry is used to capture the required information to be saved by
// the configured result logger
type TuringResultLogEntry struct {
	ctx       *context.Context
	timestamp time.Time
	request   requestLogEntry
	responses map[string]responseLogEntry
}

// NewTuringResultLogEntry returns a new TuringResultLogEntry object with the given context
// and request
func NewTuringResultLogEntry(
	ctx context.Context,
	timestamp time.Time,
	header *http.Header,
	body []byte,
) *TuringResultLogEntry {
	return &TuringResultLogEntry{
		ctx:       &ctx,
		timestamp: timestamp,
		request: requestLogEntry{
			Header: header,
			Body:   json.RawMessage(body),
		},
		responses: map[string]responseLogEntry{},
	}
}

// AddResponse adds the per-component response/error info to the TuringResultLogEntry
func (e *TuringResultLogEntry) AddResponse(key string, response []byte, err string) {
	// Check if the key supplied is valid
	if key == ResultLogKeys.Experiment ||
		key == ResultLogKeys.Enricher ||
		key == ResultLogKeys.Router ||
		key == ResultLogKeys.Ensembler {
		e.responses[key] = responseLogEntry{
			Response: json.RawMessage(response),
			Error:    err,
		}
	}
}

// TuringResultLogger is an abstraction for the underlying result logger
type TuringResultLogger interface {
	write(*TuringResultLogEntry) error
}

// InitTuringResultLogger initializes the global logger to the appropriate logger for
// recording the turing request summary
func InitTuringResultLogger(cfg *config.AppConfig) error {
	var err error

	switch cfg.ResultLogger {
	case config.BigqueryLogger:
		var bqLogger BigQueryLogger
		log.Glob().Info("Initializing BigQuery Result Logger")
		bqLogger, err = newBigQueryLogger(cfg.Name, cfg.BigQuery)
		if err != nil {
			return err
		}

		// Check if streaming insert or batch logging
		if cfg.BigQuery.BatchLoad {
			// Init fluentd logger for batch logging
			globalLogger, err = newFluentdLogger(cfg.Fluentd, bqLogger)
		} else {
			// Use BigQueryLogger for streaming insert
			globalLogger = bqLogger
		}
	case config.ConsoleLogger:
		log.Glob().Info("Initializing Console Result Logger")
		globalLogger = newConsoleLogger()
	case config.KafkaLogger:
		log.Glob().Info("Initializing Kafka Result Logger")
		globalLogger, err = newKafkaLogger(cfg.Name, cfg.Kafka)
	case config.NopLogger:
		log.Glob().Info("Initializing Nop Result Logger")
		globalLogger = newNopLogger()
	default:
		err = errors.Newf(errors.BadInput, "Unrecognized Result Logger: %s", cfg.ResultLogger)
	}

	return err
}

// LogEntry sends the input TuringResultLogEntry to the appropriate logger
func LogEntry(turLogEntry *TuringResultLogEntry) error {
	return globalLogger.write(turLogEntry)
}

func getTuringResponseOrNil(
	responseMap map[string]responseLogEntry,
	key string,
) *responseLogEntry {
	if value, ok := responseMap[key]; ok {
		return &value
	}
	return nil
}
