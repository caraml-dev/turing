package resultlog

import (
	"encoding/json"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/gojek/turing/engines/router/missionctl/log/resultlog/proto/turing"
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
	appName       string
	serialization config.SerializationFormat
	topic         string
	producer      kafkaProducer
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
		appName:       appName,
		serialization: cfg.Serialization,
		topic:         cfg.Topic,
		producer:      producer,
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

	// Format Kafka Message
	var keyBytes, valueBytes []byte
	if l.serialization == config.JsonSerializationFormat {
		valueBytes, err = newJSONKafkaLogEntry(l.appName, turLogEntry)
	} else if l.serialization == config.ProtobufSerializationFormat {
		keyBytes, valueBytes, err = newProtobufKafkaLogEntry(l.appName, turLogEntry)
	}
	if err != nil {
		return err
	}

	// Produce Message
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)
	err = l.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &l.topic,
			Partition: kafka.PartitionAny},
		Value: valueBytes,
		Key:   keyBytes,
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

// newJSONKafkaLogEntry converts a given TuringResultLogEntry to arbitrary struct with the
// required fields and marshals it into a byte array, for writing to a Kafka topic
func newJSONKafkaLogEntry(routerVersion string, e *TuringResultLogEntry) (messageBytes []byte, err error) {
	// Get Turing request id
	var turingReqID string
	turingReqID, err = turingctx.GetRequestID(*e.ctx)
	if err != nil {
		return nil, err
	}

	kafkaMessage := &struct {
		RouterVersion  string            `json:"router_version"`
		TuringReqID    string            `json:"turing_req_id"`
		EventTimestamp string            `json:"event_timestamp"`
		Request        requestLogEntry   `json:"request"`
		Experiment     *responseLogEntry `json:"experiment,omitempty"`
		Enricher       *responseLogEntry `json:"enricher,omitempty"`
		Router         *responseLogEntry `json:"router,omitempty"`
		Ensembler      *responseLogEntry `json:"ensembler,omitempty"`
	}{
		RouterVersion:  routerVersion,
		TuringReqID:    turingReqID,
		EventTimestamp: e.timestamp.Format(time.RFC3339Nano),
		Request:        e.request,
		Experiment:     getTuringResponseOrNil(e.responses, ResultLogKeys.Experiment),
		Enricher:       getTuringResponseOrNil(e.responses, ResultLogKeys.Enricher),
		Router:         getTuringResponseOrNil(e.responses, ResultLogKeys.Router),
		Ensembler:      getTuringResponseOrNil(e.responses, ResultLogKeys.Ensembler),
	}

	messageBytes, err = json.Marshal(kafkaMessage)
	if err != nil {
		return nil, err
	}
	return
}

// newProtobufKafkaLogEntry converts a given TuringResultLogEntry to the Protbuf format and marshals it,
// for writing to a Kafka topic
func newProtobufKafkaLogEntry(
	routerVersion string,
	e *TuringResultLogEntry,
) (keyBytes []byte, valueBytes []byte, err error) {
	// Get Turing request id
	var turingReqID string
	turingReqID, err = turingctx.GetRequestID(*e.ctx)
	if err != nil {
		return nil, nil, err
	}

	// Format the event timestamp and the Turing request header to the expected value
	timestamp := timestamppb.New(e.timestamp)
	reqHeader := map[string]*turing.ListOfString{}
	for key, value := range *e.request.Header {
		reqHeader[key] = &turing.ListOfString{
			Value: value,
		}
	}

	// Create the Kafka Key and Message
	newProtobufResponse := func(logEntry *responseLogEntry) *turing.Response {
		if logEntry == nil {
			return nil
		}
		return &turing.Response{
			Response: string(logEntry.Response),
			Error:    logEntry.Error,
		}
	}
	key := &turing.TuringResultLogKey{
		TuringReqId:    turingReqID,
		EventTimestamp: timestamp,
	}
	message := &turing.TuringResultLogMessage{
		RouterVersion:  routerVersion,
		TuringReqId:    turingReqID,
		EventTimestamp: timestamp,
		Request: &turing.Request{
			Header: reqHeader,
			Body:   string(e.request.Body),
		},
		Experiment: newProtobufResponse(getTuringResponseOrNil(e.responses, ResultLogKeys.Experiment)),
		Enricher:   newProtobufResponse(getTuringResponseOrNil(e.responses, ResultLogKeys.Enricher)),
		Router:     newProtobufResponse(getTuringResponseOrNil(e.responses, ResultLogKeys.Router)),
		Ensembler:  newProtobufResponse(getTuringResponseOrNil(e.responses, ResultLogKeys.Ensembler)),
	}

	// Marshal the key and the value
	keyBytes, err = proto.Marshal(key)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Unable to marshal log entry key")
	}
	valueBytes, err = proto.Marshal(message)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Unable to marshal log entry value")
	}
	return
}
