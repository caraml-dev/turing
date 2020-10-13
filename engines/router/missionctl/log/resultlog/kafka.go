package resultlog

import (
	"encoding/json"
	"time"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

const (
	kafkaConnectTimeoutMs = 1000
)

// kafkaProducer minimally defines the functionality used by the KafkaLogger,
// for producing messages to a Kafka topic (useful for mocking in tests).
type kafkaProducer interface {
	GetMetadata(*string, bool, int) (*kafka.Metadata, error)
	Produce(*kafka.Message, chan kafka.Event) error
}

// KafkaLogger logs the result log data to the configured Kafka topic
type KafkaLogger struct {
	appName  string
	topic    string
	producer kafkaProducer
}

// kafkaLogEntry is used to parse the TuringResultLogEntry and convert it into
// a loggable structure
type kafkaLogEntry struct {
	RouterVersion string            `json:"router_version"`
	TuringReqID   string            `json:"turing_req_id"`
	Timestamp     string            `json:"ts"`
	Request       requestLogEntry   `json:"request"`
	Experiment    *responseLogEntry `json:"experiment,omitempty"`
	Enricher      *responseLogEntry `json:"enricher,omitempty"`
	Router        *responseLogEntry `json:"router,omitempty"`
	Ensembler     *responseLogEntry `json:"ensembler,omitempty"`
}

// newKafkaLogger creates a new KafkaLogger
func newKafkaLogger(
	appName string,
	cfg *config.KafkaConfig,
) (*KafkaLogger, error) {
	// Create Kafka Producer
	producer, err := newKafkaProducer(cfg)
	if err != nil {
		return nil, err
	}
	// Test that we are able to query the broker on the topic. If the topic
	// does not already exist on the broker, this should create it.
	_, err = producer.GetMetadata(&cfg.Topic, false, kafkaConnectTimeoutMs)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Error Querying topic %s from Kafka broker(s)", cfg.Topic)
	}
	// Create Kafka Logger
	return &KafkaLogger{
		appName:  appName,
		topic:    cfg.Topic,
		producer: producer,
	}, nil
}

func newKafkaProducer(cfg *config.KafkaConfig) (kafkaProducer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Brokers})
	if err != nil {
		return nil, errors.Wrapf(err, "Error initializing Kafka Producer")
	}
	return producer, err
}

func (l *KafkaLogger) write(turLogEntry *TuringResultLogEntry) error {
	var err error

	// Measure time taken to marshal the data and write the log to the kafka topic
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "kafka_marshal_and_write"
			},
		},
	)()

	// Make the kafkaLogEntry
	var logEntry *kafkaLogEntry
	logEntry, err = newKafkaLogEntry(l.appName, turLogEntry)
	if err != nil {
		return err
	}
	// Marshal data
	var bytes []byte
	bytes, err = json.Marshal(logEntry)
	if err != nil {
		return err
	}

	// Produce message
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)
	err = l.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &l.topic,
			Partition: kafka.PartitionAny},
		Value: bytes,
	}, deliveryChan)
	if err != nil {
		return err
	}

	// Get delivery response
	event := <-deliveryChan
	msg := event.(*kafka.Message)
	if msg.TopicPartition.Error != nil {
		err = errors.Newf(errors.BadResponse,
			"Delivery failed: %v\n", msg.TopicPartition.Error)
		return err
	}

	return nil
}

// newKafkaLogEntry converts a given TuringResultLogEntry to a kafkaLogEntry, to
// be written to a Kafka topic
func newKafkaLogEntry(routerVersion string, e *TuringResultLogEntry) (*kafkaLogEntry, error) {
	// Get Turing request id
	turingReqID, err := turingctx.GetRequestID(*e.ctx)
	if err != nil {
		return nil, err
	}

	logEntry := &kafkaLogEntry{
		RouterVersion: routerVersion,
		TuringReqID:   turingReqID,
		Timestamp:     e.timestamp.Format(time.RFC3339Nano),
		Request:       e.request,
		Experiment:    getTuringResponseOrNil(e.responses, ResultLogKeys.Experiment),
		Enricher:      getTuringResponseOrNil(e.responses, ResultLogKeys.Enricher),
		Router:        getTuringResponseOrNil(e.responses, ResultLogKeys.Router),
		Ensembler:     getTuringResponseOrNil(e.responses, ResultLogKeys.Ensembler),
	}
	// Add optional fields
	return logEntry, nil
}
