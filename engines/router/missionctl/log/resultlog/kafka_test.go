package resultlog

import (
	"encoding/json"
	"net/http"
	"testing"

	"bou.ke/monkey"
	"github.com/gojek/turing/engines/router/missionctl/config"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
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
	monkey.Patch(newKafkaProducer, func(cfg *config.KafkaConfig) (kafkaProducer, error) {
		return mockProducer, nil
	})
	defer monkey.Unpatch(newKafkaProducer)

	// Set up GetMetadata on the mock producer
	mockProducer.On("GetMetadata", &cfg.Topic, false, 1000).Return(nil, nil)

	// Create the new logger and validate
	testLogger, err := newKafkaLogger("test-app-name", &cfg)
	assert.NoError(t, err)
	assert.Equal(t, *testLogger, KafkaLogger{
		appName:  "test-app-name",
		topic:    cfg.Topic,
		producer: mockProducer,
	})
	mockProducer.AssertCalled(t, "GetMetadata", &cfg.Topic, false, 1000)
}

func TestNewKafkaLogEntry(t *testing.T) {
	// Create test Turing log entry
	ctx, turingLogEntry := makeTestTuringResultLogEntry(t)

	// Run newKafkaLogEntry and validate
	logEntry, err := newKafkaLogEntry("test-app-name", turingLogEntry)
	assert.NoError(t, err)
	// Get Turing request id
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)
	// Set the Timestamp on logEntry, before comparison
	logEntry.Timestamp = "test-timestamp"
	// Compare logEntry data
	assert.Equal(t, kafkaLogEntry{
		RouterVersion: "test-app-name",
		TuringReqID:   turingReqID,
		Timestamp:     "test-timestamp",
		Request: requestLogEntry{
			Header: &http.Header{
				"Req_id": []string{"test_req_id"},
			},
			Body: []byte(`{"customer_id": "test_customer"}`),
		},
		Experiment: &responseLogEntry{
			Response: []byte(`{"key": "experiment_data"}`),
			Error:    "",
		},
		Router: &responseLogEntry{
			Response: []byte(`{"key": "router_data"}`),
			Error:    "",
		},
		Enricher: &responseLogEntry{
			Response: []byte(`{"key": "enricher_data"}`),
			Error:    "",
		},
		Ensembler: nil,
	}, *logEntry)
}

func TestKafkaLoggerWrite(t *testing.T) {
	// Create test logger and log entry
	mp := &mockKafkaProducer{}
	logger := &KafkaLogger{
		appName:  "test-app-name",
		topic:    "test-topic",
		producer: mp,
	}
	testKafkaLogEntry := &kafkaLogEntry{}
	turingResLogEntry := &TuringResultLogEntry{}

	// Patch newKafkaLogEntry
	monkey.Patch(
		newKafkaLogEntry,
		func(routerVersion string, entry *TuringResultLogEntry) (*kafkaLogEntry, error) {
			// Test that the function is called with the expected args
			assert.Equal(t, "test-app-name", routerVersion)
			assert.Equal(t, turingResLogEntry, entry)
			return testKafkaLogEntry, nil
		},
	)
	defer monkey.UnpatchAll()

	// Set up Produce
	mp.On("Produce", mock.Anything, mock.Anything).Return(nil)

	// Call write and assert that Produce is called with the expected arguments
	err := logger.write(turingResLogEntry)
	assert.NoError(t, err)
	// Marshal the Kafka log entry
	byteData, err := json.Marshal(testKafkaLogEntry)
	tu.FailOnError(t, err)
	mp.AssertCalled(t, "Produce", &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &logger.topic,
			Partition: kafka.PartitionAny},
		Value: byteData,
	}, mock.Anything)
}
