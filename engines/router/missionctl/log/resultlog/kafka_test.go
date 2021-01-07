package resultlog

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"testing"

// 	"bou.ke/monkey"
// 	"github.com/gojek/turing/engines/router/missionctl/config"
// 	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
// 	"github.com/gojek/turing/engines/router/missionctl/turingctx"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
// )

// // mockKafkaProducer implements the kafkaProducer
// type mockKafkaProducer struct {
// 	mock.Mock
// }

// func (mp *mockKafkaProducer) GetMetadata(
// 	topic *string,
// 	allTopics bool,
// 	timeoutMs int,
// ) (*kafka.Metadata, error) {
// 	mp.Called(topic, allTopics, timeoutMs)
// 	return nil, nil
// }

// func (mp *mockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
// 	mp.Called(msg, deliveryChan)
// 	// Send event to deliveryChan
// 	deliveryChan <- &kafka.Message{}
// 	return nil
// }

// func TestNewKafkaProducer(t *testing.T) {
// 	// Patch the kafka.NewProducer method to validate input
// 	cfg := &config.KafkaConfig{
// 		Brokers: "localhost:9090",
// 		Topic:   "kafka_topic",
// 	}
// 	monkey.Patch(kafka.NewProducer,
// 		func(conf *kafka.ConfigMap) (*kafka.Producer, error) {
// 			// Check that the expected config is passed in
// 			bootstapServers, err := conf.Get("bootstrap.servers", "")
// 			assert.NoError(t, err)
// 			assert.Equal(t, cfg.Brokers, bootstapServers)
// 			return nil, nil
// 		},
// 	)
// 	defer monkey.Unpatch(kafka.NewProducer)
// 	// Test newKafkaProducer
// 	_, err := newKafkaProducer(cfg)
// 	assert.NoError(t, err)
// }

// func TestNewKafkaLogger(t *testing.T) {
// 	// Make test config
// 	cfg := config.KafkaConfig{
// 		Brokers: "localhost:9001",
// 		Topic:   "kafka_topic",
// 	}

// 	// Patch the newKafkaProducer method to return the mock producer
// 	mockProducer := &mockKafkaProducer{}
// 	monkey.Patch(newKafkaProducer, func(cfg *config.KafkaConfig) (kafkaProducer, error) {
// 		return mockProducer, nil
// 	})
// 	defer monkey.Unpatch(newKafkaProducer)

// 	// Set up GetMetadata on the mock producer
// 	mockProducer.On("GetMetadata", &cfg.Topic, false, 1000).Return(nil, nil)

// 	// Create the new logger and validate
// 	testLogger, err := newKafkaLogger("test-app-name", &cfg)
// 	assert.NoError(t, err)
// 	assert.Equal(t, *testLogger, KafkaLogger{
// 		appName:  "test-app-name",
// 		topic:    cfg.Topic,
// 		producer: mockProducer,
// 	})
// 	mockProducer.AssertCalled(t, "GetMetadata", &cfg.Topic, false, 1000)
// }

// func TestNewJSONKafkaLogEntry(t *testing.T) {
// 	// Create test Turing log entry
// 	ctx, turingLogEntry := makeTestTuringResultLogEntry(t)

// 	// Run newJSONKafkaLogEntry and validate
// 	message, err := newJSONKafkaLogEntry("test-app-name", turingLogEntry)
// 	assert.NoError(t, err)
// 	// Get Turing request id
// 	turingReqID, err := turingctx.GetRequestID(ctx)
// 	tu.FailOnError(t, err)
// 	// Compare logEntry data
// 	assert.JSONEq(t,
// 		fmt.Sprintf(`{
// 			"router_version": "test-app-name",
// 			"turing_req_id": "%s",
// 			"event_timestamp": "2000-02-01T04:05:06.000000007Z",
// 			"request": {
// 				"header": {"Req_id": "test_req_id"},
// 				"body": "{\"customer_id\": \"test_customer\"}"
// 			},
// 			"experiment": {
// 				"error": "Error received"
// 			},
// 			"enricher": {
// 				"response": "{\"key\": \"enricher_data\"}"
// 			},
// 			"router": {
// 				"response": "{\"key\": \"router_data\"}"
// 			}
// 		}`, turingReqID),
// 		string(message),
// 	)
// }

// func TestNewProtobufKafkaLogEntry(t *testing.T) {
// 	// Create test Turing log entry
// 	_, turingLogEntry := makeTestTuringResultLogEntry(t)
// 	// Overwrite context field with turing request id value
// 	*turingLogEntry.ctx = turingctx.NewTestTuringContext(context.Background(), "testID")
// 	// Run newProtobufKafkaLogEntry and validate
// 	key, message, err := newProtobufKafkaLogEntry("test-app-name", turingLogEntry)
// 	assert.NoError(t, err)
// 	// Compare logEntry data
// 	assert.Equal(t, "\n\x06testID\x12\b\b\xf2\xb6\xd9\xc4\x03\x10\a", string(key))
// 	assert.Equal(t, strings.Join([]string{
// 		"\n\x06testID\x12\b\b\xf2\xb6\xd9\xc4\x03\x10\a\x1a\rtest-app-name\"9\n\x15\n\x06",
// 		"Req_id\x12\vtest_req_id\x12 {\"customer_id\": \"test_customer\"}*\x10\x12\x0e",
// 		"Error received2\x1a\n\x18{\"key\": \"enricher_data\"}:\x18\n\x16{\"key\": \"router_data\"}",
// 	}, ""), string(message))
// }

// func TestKafkaLoggerWrite(t *testing.T) {
// 	// Create test logger and log entry
// 	mp := &mockKafkaProducer{}
// 	logger := &KafkaLogger{
// 		appName:             "test-app-name",
// 		serializationFormat: "json",
// 		topic:               "test-topic",
// 		producer:            mp,
// 	}
// 	testKafkaLogEntry := []byte(`{"key": "value"}`)
// 	turingResLogEntry := &TuringResultLogEntry{}

// 	// Patch newKafkaLogEntry
// 	monkey.Patch(
// 		newJSONKafkaLogEntry,
// 		func(routerVersion string, entry *TuringResultLogEntry) ([]byte, error) {
// 			// Test that the function is called with the expected args
// 			assert.Equal(t, "test-app-name", routerVersion)
// 			assert.Equal(t, turingResLogEntry, entry)
// 			return testKafkaLogEntry, nil
// 		},
// 	)
// 	defer monkey.UnpatchAll()

// 	// Set up Produce
// 	mp.On("Produce", mock.Anything, mock.Anything).Return(nil)

// 	// Call write and assert that Produce is called with the expected arguments
// 	err := logger.write(turingResLogEntry)
// 	assert.NoError(t, err)
// 	mp.AssertCalled(t, "Produce", &kafka.Message{
// 		TopicPartition: kafka.TopicPartition{
// 			Topic:     &logger.topic,
// 			Partition: kafka.PartitionAny},
// 		Value: testKafkaLogEntry,
// 	}, mock.Anything)
// }
