package resultlog

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/log/resultlog/proto/turing"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

// Init the global logger to Nop Logger, calling InitTuringResultLogger will reset this.
var globalLogger TuringResultLogger = newNopLogger()

// appName stores the configured app name, to be applied to each log entry
var appName string

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

// TuringResultLogEntry represents the information logged by the result logger
type TuringResultLogEntry turing.TuringResultLogMessage

// MarshalJSON implements custom Marshaling for TuringResultLogEntry, using the underlying proto def
func (logEntry *TuringResultLogEntry) MarshalJSON() ([]byte, error) {
	m := &protojson.MarshalOptions{
		UseProtoNames: true, // Use the json field name instead of the camel case struct field name
	}
	message := (*turing.TuringResultLogMessage)(logEntry)
	return m.Marshal(message)
}

// Value returns the TuringResultLogEntry in a loggable format
func (logEntry *TuringResultLogEntry) Value() (map[string]interface{}, error) {
	var kvPairs map[string]interface{}
	// Marshal into bytes
	bytes, err := json.Marshal(&logEntry)
	if err != nil {
		return kvPairs, errors.Wrapf(err, "Error marshaling the result log")
	}
	// Unmarshal into map[string]interface{}
	err = json.Unmarshal(bytes, &kvPairs)
	if err != nil {
		return kvPairs, errors.Wrapf(err, "Error unmarshaling the result log")
	}
	return kvPairs, nil
}

// AddResponse adds the per-component response/error info to the TuringResultLogEntry
func (logEntry *TuringResultLogEntry) AddResponse(key string, responseBody []byte, responseHeader string, err string) {
	responseRecord := &turing.Response{
		ResponseHeader: responseHeader,
		ResponseBody:   string(json.RawMessage(responseBody)),
		Error:          err,
	}
	switch key {
	case ResultLogKeys.Experiment:
		logEntry.Experiment = responseRecord
	case ResultLogKeys.Enricher:
		logEntry.Enricher = responseRecord
	case ResultLogKeys.Router:
		logEntry.Router = responseRecord
	case ResultLogKeys.Ensembler:
		logEntry.Ensembler = responseRecord
	}
}

// NewTuringResultLogEntry returns a new TuringResultLogEntry object with the given context
// and request
func NewTuringResultLogEntry(
	ctx context.Context,
	timestamp time.Time,
	header *http.Header,
	body []byte,
) *TuringResultLogEntry {
	// Get Turing Request Id
	turingReqID, _ := turingctx.GetRequestID(ctx)

	// Format Request Header
	reqHeader := map[string]string{}
	for k, v := range *header {
		reqHeader[k] = strings.Join(v, ",")
	}

	return &TuringResultLogEntry{
		TuringReqId:    turingReqID,
		EventTimestamp: timestamppb.New(timestamp),
		RouterVersion:  appName,
		Request: &turing.Request{
			Header: reqHeader,
			Body:   string(json.RawMessage(body)),
		},
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

	// Save the configured app name to the package var
	appName = cfg.Name

	switch cfg.ResultLogger {
	case config.BigqueryLogger:
		var bqLogger BigQueryLogger
		log.Glob().Info("Initializing BigQuery Result Logger")
		// Init BQ logger. This will also run the necessary checks on the table schema /
		// create it if not exists.
		bqLogger, err = newBigQueryLogger(cfg.BigQuery)
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
		globalLogger, err = newKafkaLogger(cfg.Kafka)
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
