package resultlog

import (
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
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
	appName             string
	serializationFormat config.SerializationFormat
	topic               string
	producer            kafkaProducer
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
		appName:             appName,
		serializationFormat: cfg.SerializationFormat,
		topic:               cfg.Topic,
		producer:            producer,
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
	if l.serializationFormat == config.JSONSerializationFormat {
		valueBytes, err = newJSONKafkaLogEntry(l.appName, turLogEntry)
	} else if l.serializationFormat == config.ProtobufSerializationFormat {
		keyBytes, valueBytes, err = newProtobufKafkaLogEntry(l.appName, turLogEntry)
	} else {
		// Unknown format, we wouldn't hit this since the config is checked at initialization,
		// but handle it.
		return errors.Newf(errors.BadConfig, "Unknown Serialization format %s", l.serializationFormat)
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

// newJSONKafkaLogEntry converts a given TuringResultLogEntry to a JSON message of the
// TuringResultLogMessage type, for writing to a Kafka topic
func newJSONKafkaLogEntry(
	routerVersion string,
	resultLogEntry *TuringResultLogEntry,
) (messageBytes []byte, err error) {
	// Get Turing request ID
	var turingReqID string
	turingReqID, err = turingctx.GetRequestID(*resultLogEntry.ctx)
	if err != nil {
		return nil, err
	}

	// Create the Kafka message
	kafkaMessage, err := newTuringResultLogMessage(routerVersion, resultLogEntry, turingReqID)
	if err != nil {
		return nil, err
	}

	// Marshal to JSON
	m := &protojson.MarshalOptions{
		UseProtoNames: true, // Use the json field name instead of the camel case struct field name
	}
	messageBytes, err = m.Marshal(kafkaMessage)
	if err != nil {
		return nil, err
	}
	return
}

// newProtobufKafkaLogEntry converts a given TuringResultLogEntry to the Protbuf format and marshals it,
// for writing to a Kafka topic
func newProtobufKafkaLogEntry(
	routerVersion string,
	resultLogEntry *TuringResultLogEntry,
) (keyBytes []byte, valueBytes []byte, err error) {
	// Get Turing request ID
	var turingReqID string
	turingReqID, err = turingctx.GetRequestID(*resultLogEntry.ctx)
	if err != nil {
		return nil, nil, err
	}

	// Create the Kafka key and message
	key := &turing.TuringResultLogKey{
		TuringReqId:    turingReqID,
		EventTimestamp: timestamppb.New(resultLogEntry.timestamp),
	}
	message, err := newTuringResultLogMessage(routerVersion, resultLogEntry, turingReqID)
	if err != nil {
		return nil, nil, err
	}

	// Marshal the key and the message
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

func newTuringResultLogMessage(
	routerVersion string,
	resultLogEntry *TuringResultLogEntry,
	turingReqID string,
) (*turing.TuringResultLogMessage, error) {
	// Format the Turing request header per the proto definition
	reqHeader := map[string]string{}
	for key, values := range *resultLogEntry.request.Header {
		reqHeader[key] = strings.Join(values, ",")
	}

	// Create the Kafka Message
	newProtobufResponse := func(e *responseLogEntry) *turing.Response {
		if e == nil {
			return nil
		}
		if e.Response == nil {
			return &turing.Response{
				Error: e.Error,
			}
		}
		// Format response body as string
		return &turing.Response{
			Response: string(e.Response),
			Error:    e.Error,
		}
	}
	message := &turing.TuringResultLogMessage{
		RouterVersion:  routerVersion,
		TuringReqId:    turingReqID,
		EventTimestamp: timestamppb.New(resultLogEntry.timestamp),
		Request: &turing.Request{
			Header: reqHeader,
			Body:   string(resultLogEntry.request.Body),
		},
		Experiment: newProtobufResponse(getTuringResponseOrNil(resultLogEntry.responses, ResultLogKeys.Experiment)),
		Enricher:   newProtobufResponse(getTuringResponseOrNil(resultLogEntry.responses, ResultLogKeys.Enricher)),
		Router:     newProtobufResponse(getTuringResponseOrNil(resultLogEntry.responses, ResultLogKeys.Router)),
		Ensembler:  newProtobufResponse(getTuringResponseOrNil(resultLogEntry.responses, ResultLogKeys.Ensembler)),
	}

	return message, nil
}
