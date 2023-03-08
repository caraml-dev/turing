package resultlog

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
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
	serializationFormat config.SerializationFormat
	topic               string
	producer            kafkaProducer
}

// NewKafkaLogger creates a new KafkaLogger
func NewKafkaLogger(cfg *config.KafkaConfig) (*KafkaLogger, error) {
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
		serializationFormat: cfg.SerializationFormat,
		topic:               cfg.Topic,
		producer:            producer,
	}, nil
}

func newKafkaProducer(cfg *config.KafkaConfig) (kafkaProducer, error) {
	producer, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": cfg.Brokers,
			"message.max.bytes": cfg.MaxMessageBytes,
			"compression.type":  cfg.CompressionType})
	if err != nil {
		return nil, errors.Wrapf(err, "Error initializing Kafka Producer")
	}
	return producer, err
}

func (l *KafkaLogger) writeToKafka(
	message proto.Message,
	turingReqID string,
	timestamp *timestamppb.Timestamp) error {
	var err error

	// Measure time taken to marshal the data and write the log to the kafka topic
	defer metrics.Glob().MeasureDurationMs(
		instrumentation.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "kafka_marshal_and_write"
			},
			"traffic_rule": func() string { return "" },
		},
	)()

	// Format Kafka Message
	var keyBytes, valueBytes []byte
	if l.serializationFormat == config.JSONSerializationFormat {
		valueBytes, err = newJSONKafkaLogEntry(message)
	} else if l.serializationFormat == config.ProtobufSerializationFormat {
		keyBytes, valueBytes, err = newProtobufKafkaLogEntry(
			message,
			turingReqID,
			timestamp)
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

func (l *KafkaLogger) write(turLogEntry *turing.TuringResultLogMessage) error {
	return l.writeToKafka(
		turLogEntry,
		turLogEntry.TuringReqId,
		turLogEntry.EventTimestamp,
	)
}

// newJSONKafkaLogEntry converts a given TuringResultLogEntry to  bytes, for writing to a Kafka topic
// in JSON format
func newJSONKafkaLogEntry(message proto.Message) (messageBytes []byte, err error) {
	messageBytes, err = protoJSONMarshaller.Marshal(message)
	if err != nil {
		return nil, err
	}
	return
}

// newProtobufKafkaLogEntry converts a given TuringResultLogEntry to the Protobuf format and marshals it,
// for writing to a Kafka topic
func newProtobufKafkaLogEntry(
	message proto.Message,
	turingReqID string,
	eventTimestamp *timestamppb.Timestamp,
) (keyBytes []byte, valueBytes []byte, err error) {
	// Create the Kafka key
	key := &turing.TuringResultLogKey{
		TuringReqId:    turingReqID,
		EventTimestamp: eventTimestamp,
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
