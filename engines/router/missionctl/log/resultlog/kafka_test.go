package resultlog

import (
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// mockKafkaProducer implements the kafkaProducer
type mockKafkaProducer struct {
	mock.Mock
}

func (mp *mockKafkaProducer) GetMetadata(
	topic *string,
	allTopics bool,
	timeoutMs int,
) (*kafka.Metadata, error) {
	mp.Called(topic, allTopics, timeoutMs)
	return nil, nil
}

func (mp *mockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	mp.Called(msg, deliveryChan)
	// Send event to deliveryChan
	deliveryChan <- &kafka.Message{}
	return nil
}

func TestNewKafkaProducer(t *testing.T) {
	// Patch the kafka.NewProducer method to validate input
	cfg := &config.KafkaConfig{
		Brokers: "localhost:9090",
		Topic:   "kafka_topic",
	}
	monkey.Patch(kafka.NewProducer,
		func(conf *kafka.ConfigMap) (*kafka.Producer, error) {
			// Check that the expected config is passed in
			bootstapServers, err := conf.Get("bootstrap.servers", "")
			assert.NoError(t, err)
			assert.Equal(t, cfg.Brokers, bootstapServers)
			return nil, nil
		},
	)
	defer monkey.Unpatch(kafka.NewProducer)
	// Test newKafkaProducer
	_, err := newKafkaProducer(cfg)
	assert.NoError(t, err)
}

func TestNewKafkaLogger(t *testing.T) {
	// Make test config
	cfg := config.KafkaConfig{
		Brokers: "localhost:9001",
		Topic:   "kafka_topic",
	}

	// Patch the newKafkaProducer method to return the mock producer
	mockProducer := &mockKafkaProducer{}
	monkey.Patch(newKafkaProducer, func(_ *config.KafkaConfig) (kafkaProducer, error) {
		return mockProducer, nil
	})
	defer monkey.Unpatch(newKafkaProducer)

	// Set up GetMetadata on the mock producer
	mockProducer.On("GetMetadata", &cfg.Topic, false, 1000).Return(nil, nil)

	// Create the new logger and validate
	testLogger, err := NewKafkaLogger(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, *testLogger, KafkaLogger{
		topic:    cfg.Topic,
		producer: mockProducer,
	})
	mockProducer.AssertCalled(t, "GetMetadata", &cfg.Topic, false, 1000)
}

func TestNewJSONKafkaLogEntry(t *testing.T) {
	// Create test Turing log entry
	ctx, turingLogEntry := makeTestTuringResultLog(t)

	// Run newJSONKafkaLogEntry and validate
	message, err := newJSONKafkaLogEntry(turingLogEntry)
	assert.NoError(t, err)
	// Get Turing request id
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)
	// Compare logEntry data
	assert.JSONEq(t,
		fmt.Sprintf(`{
			"turing_req_id": "%s",
			"event_timestamp": "2000-02-01T04:05:06.000000007Z",
			"request": {
				"header": {
					"Req_id": "test_req_id"
				},
				"body": "{\"customer_id\": \"test_customer\"}"
			},
			"experiment": {
				"error": "Error received"
			},
			"enricher":{
				"response":"{\"key\": \"enricher_data\"}", 
				"header":{"Content-Encoding":"lz4","Content-Type":"text/html,charset=utf-8"}
			},
			"router":{
				"response":"{\"key\": \"router_data\"}",
				"header":{"Content-Encoding":"gzip","Content-Type":"text/html,charset=utf-8"}
			}
		}`, turingReqID),
		string(message),
	)
}

func TestNewProtobufKafkaLogEntry(t *testing.T) {
	// Create test Turing log entry
	_, turingLogEntry := makeTestTuringResultLog(t)
	resultLogMessage := turingLogEntry
	// Overwrite the turing request id value
	resultLogMessage.TuringReqId = "testID"

	// Run newProtobufKafkaLogEntry and validate
	key, message, err := newProtobufKafkaLogEntry(
		resultLogMessage,
		resultLogMessage.TuringReqId,
		resultLogMessage.EventTimestamp)
	assert.NoError(t, err)

	// Unmarshall serialised message
	decodedTuringResultLogMessage := &turing.TuringResultLogMessage{}
	err = proto.Unmarshal(message, decodedTuringResultLogMessage)
	assert.NoError(t, err)
	// Convert expected and actual log entries to JSON for comparison
	expectedJSON, err := protoJSONMarshaller.Marshal(turingLogEntry)
	assert.NoError(t, err)
	m := protojson.MarshalOptions{
		UseProtoNames: true,
	}
	actualJSON, err := m.Marshal(decodedTuringResultLogMessage)
	assert.NoError(t, err)

	// Compare logEntry data
	assert.Equal(t, "\n\x06testID\x12\b\b\xf2\xb6\xd9\xc4\x03\x10\a", string(key))
	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

func TestKafkaLoggerWrite(t *testing.T) {
	// Create test logger and log entry
	mp := &mockKafkaProducer{}
	logger := &KafkaLogger{
		serializationFormat: "json",
		topic:               "test-topic",
		producer:            mp,
	}
	testKafkaLogEntry := []byte(`{"turing_req_id":"123"}`)
	msg := &turing.TuringResultLogMessage{TuringReqId: "123"}
	expectedMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &logger.topic,
			Partition: kafka.PartitionAny},
		Value: testKafkaLogEntry,
	}

	// Set up Produce
	mp.On("Produce", expectedMessage, mock.Anything).Return(nil)

	// Call write and assert that Produce is called with the expected arguments
	err := logger.write(msg)
	assert.NoError(t, err)
	mp.AssertCalled(t, "Produce", expectedMessage, mock.Anything)
}
