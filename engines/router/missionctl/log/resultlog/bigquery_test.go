// +build integration

package resultlog

/*
Some tests in this file are integration tests that exercise the BigQuery client
and will only run with go test --tags=integration. For these tests
to work, a GCP service account key with the right access must be set in the
environment where the tests are run. The BigQuery dataset and project being tested
must also exist.
*/

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/turing/engines/router/missionctl/config"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

type testSuiteBQSchema struct {
	filepath1 string
	filepath2 string
	isUpdated bool
	isError   bool
}

func TestGetTuringResultTableSchema(t *testing.T) {
	// Get expected schema
	bytes, err := tu.ReadFile(filepath.Join("..", "..", "testdata", "bq_turing_result_schema.json"))
	tu.FailOnError(t, err)
	expectedSchema, err := bigquery.SchemaFromJSON(bytes)
	tu.FailOnError(t, err)

	// Actual schema
	schema := getTuringResultTableSchema()

	// Enclose schema in a struct for go-cmp
	type bqSchema struct {
		Schema bigquery.Schema
	}
	wantSchema := &bqSchema{Schema: expectedSchema}
	gotSchema := &bqSchema{Schema: *schema}

	// Compare all fields except Description
	opt := cmpopts.IgnoreFields(bigquery.FieldSchema{}, "Description")
	if !cmp.Equal(wantSchema, gotSchema, opt) {
		t.Log(cmp.Diff(wantSchema, gotSchema, opt))
		t.Fail()
	}
}

func TestCheckTableSchema(t *testing.T) {
	// Test cases
	tests := map[string]testSuiteBQSchema{
		"order_diff": {
			filepath1: filepath.Join("..", "..", "testdata", "bq_schema_1_order_diff.json"),
			filepath2: filepath.Join("..", "..", "testdata", "bq_schema_1_original.json"),
			isError:   false,
		},
		"field_diff": {
			filepath1: filepath.Join("..", "..", "testdata", "bq_schema_2_field_diff.json"),
			filepath2: filepath.Join("..", "..", "testdata", "bq_schema_1_original.json"),
			isError:   true,
		},
		"required_diff": {
			filepath1: filepath.Join("..", "..", "testdata", "bq_schema_3_required_diff.json"),
			filepath2: filepath.Join("..", "..", "testdata", "bq_schema_1_original.json"),
			isError:   true,
		},
		"nested_schema_diff": {
			filepath1: filepath.Join("..", "..", "testdata", "bq_schema_4_nested_schema_diff.json"),
			filepath2: filepath.Join("..", "..", "testdata", "bq_schema_1_original.json"),
			isUpdated: true,
			isError:   false,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Read in the JSON schema from the two files
			filebytes1, _ := tu.ReadFile(data.filepath1)
			filebytes2, _ := tu.ReadFile(data.filepath2)

			// Create BQ schema
			schema1, _ := bigquery.SchemaFromJSON(filebytes1)
			schema2, _ := bigquery.SchemaFromJSON(filebytes2)

			// Compare and check the success state
			newSchema, isUpdated, err := compareTableSchema(&schema1, &schema2)
			assert.Equal(t, data.isError, err != nil)
			assert.Equal(t, data.isUpdated, isUpdated)
			// If updated, check that the new schema and the expected schema match
			if isUpdated {
				_, isUpdated, err = compareTableSchema(&schema1, newSchema)
				assert.NoError(t, err)
				assert.False(t, isUpdated)
			}
		})
	}
}

func TestBigQueryLogSave(t *testing.T) {
	// Create test Turing log entry
	ctx, turingLogEntry := makeTestTuringResultLogEntry(t)
	// Create BQ Log Entry
	logEntry := &bqLogEntry{turingLogEntry}
	// Exercise BQ Log Entry's Save method
	message, _, err := logEntry.Save()
	assert.NoError(t, err)
	// Get Turing request id
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)
	// Compare logEntry data
	assert.Equal(t, map[string]bigquery.Value{
		"router_version":  "test-app-name",
		"turing_req_id":   turingReqID,
		"event_timestamp": "2000-02-01T04:05:06.000000007Z",
		"request": bigquery.Value(struct {
			Header []bigquery.Value
			Body   bigquery.Value
		}{
			Header: []bigquery.Value{bigquery.Value(struct {
				Key   string
				Value string
			}{
				Key:   "Req_id",
				Value: "test_req_id",
			})},
			Body: bigquery.Value("{\"customer_id\": \"test_customer\"}"),
		}),
		"experiment": bigquery.Value(map[string]interface{}{
			"error": "Error received",
		}),
		"enricher": bigquery.Value(map[string]interface{}{
			"response": "{\"key\": \"enricher_data\"}",
		}),
		"router": bigquery.Value(map[string]interface{}{
			"response": "{\"key\": \"router_data\"}",
		}),
	}, message)
}

// This test case initializes the BQ client to connect to the specified
// project and dataset, where a turing result table of the given name
// does not exist. It is assumed that the environment that the test
// runs in has the required privileges.
func TestNewBigQueryLogger(t *testing.T) {
	// Create test config with a random unique table name
	cfg := &config.BQConfig{
		Project: "gcp-project-id",
		Dataset: "dataset_id",
		Table:   fmt.Sprintf("turing_test_%d", time.Now().UnixNano()),
	}

	// Init BQ Logger
	logger, err := newBigQueryLogger(cfg)
	tu.FailOnError(t, err)

	// Test logger attributes
	bqLogger, ok := logger.(*bigQueryLogger)
	if !ok {
		tu.FailOnError(t, fmt.Errorf("Unexpected data type returned"))
	}
	assert.Equal(t, cfg.Dataset, bqLogger.dataset)
	assert.Equal(t, cfg.Table, bqLogger.table)

	// Test that the newly created table has the expected schema
	expectedSchema := getTuringResultTableSchema()
	schema := getTableSchema(bqLogger.bqClient, cfg.Dataset, cfg.Table, t)
	_, isUpdated, err := compareTableSchema(&schema, expectedSchema)
	tu.FailOnError(t, err)
	assert.False(t, isUpdated)

	// Remove the newly created table
	err = deleteBigQueryTable(bqLogger.bqClient, cfg.Dataset, cfg.Table)
	assert.NoError(t, err)
}

// This test case creates a bew BQ table and then initializes the BQ client
// which is expected to update the schema
func TestNewBigQueryLoggerAddColumns(t *testing.T) {
	// Create test config with a random unique table name
	cfg := &config.BQConfig{
		Project: "gcp-project-id",
		Dataset: "dataset_id",
		Table:   fmt.Sprintf("turing_test_%d", time.Now().UnixNano()),
	}

	// Create the BQ table with reduced columns
	initialSchema := &bigquery.Schema{
		{Name: "turing_req_id", Type: bigquery.StringFieldType, Required: false},
		{Name: "ts", Type: bigquery.TimestampFieldType, Required: false},
		{Name: "router_version", Type: bigquery.StringFieldType, Required: false},
		{Name: "request", Type: bigquery.RecordFieldType,
			Required: false,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "header", Type: bigquery.StringFieldType},
			},
		},
	}
	createBQTable(t, cfg, initialSchema)

	// Init BQ Logger
	logger, err := newBigQueryLogger(cfg)
	tu.FailOnError(t, err)

	// Test logger attributes
	bqLogger, ok := logger.(*bigQueryLogger)
	if !ok {
		tu.FailOnError(t, fmt.Errorf("Unexpected data type returned"))
	}
	assert.Equal(t, cfg.Dataset, bqLogger.dataset)
	assert.Equal(t, cfg.Table, bqLogger.table)

	// Test that the newly created table has the expected schema
	expectedSchema := getTuringResultTableSchema()
	schema := getTableSchema(bqLogger.bqClient, cfg.Dataset, cfg.Table, t)
	_, isUpdated, err := compareTableSchema(&schema, expectedSchema)
	tu.FailOnError(t, err)
	assert.False(t, isUpdated)

	// Remove the newly created table
	err = deleteBigQueryTable(bqLogger.bqClient, cfg.Dataset, cfg.Table)
	assert.NoError(t, err)
}

func TestBigQueryLoggerWrite(t *testing.T) {
	// Create test BQ config with a random unique table name
	cfg := &config.BQConfig{
		Project: "gcp-project-id",
		Dataset: "dataset_id",
		Table:   fmt.Sprintf("turing_test_%d", time.Now().UnixNano()),
	}

	// Init BQ Client
	logger, err := newBigQueryLogger(cfg)
	tu.FailOnError(t, err)
	bqLogger, ok := logger.(*bigQueryLogger)
	if !ok {
		tu.FailOnError(t, fmt.Errorf("Unexpected data type returned"))
	}

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := ioutil.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Create a TuringResultLogEntry record and add the data
	entry := NewTuringResultLogEntry(ctx, time.Now(), &req.Header, reqBody)
	entry.RouterVersion = "turing-router-1"
	entry.AddResponse("experiment", []byte(`{"key": "experiment_data"}`), "")
	entry.AddResponse("enricher", []byte(`{"key": "enricher_data"}`), "")
	entry.AddResponse("router", []byte(`{"key": "router_data"}`), "")
	entry.AddResponse("ensembler", nil, "Error Response")

	// Write the log and check that there is no error
	err = LogEntry(entry)
	assert.NoError(t, err)

	// Remove newly created table
	err = deleteBigQueryTable(bqLogger.bqClient, cfg.Dataset, cfg.Table)
	assert.NoError(t, err)
}

// deleteBigQueryTable assumes that the table exists
func deleteBigQueryTable(b *bigquery.Client, datasetID, tableID string) error {
	table := b.Dataset(datasetID).Table(tableID)
	return table.Delete(context.Background())
}

// getTableSchema assumes that the table exists
func getTableSchema(b *bigquery.Client, datasetID, tableID string, t *testing.T) bigquery.Schema {
	table := b.Dataset(datasetID).Table(tableID)
	metadata, err := table.Metadata(context.Background())
	tu.FailOnError(t, err)

	return metadata.Schema
}

// createBQTable assumes that a table does not already exist and creates it with
// the given schema
func createBQTable(t *testing.T, cfg *config.BQConfig, schema *bigquery.Schema) {
	// Init BQ Client
	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, cfg.Project)
	tu.FailOnError(t, err)

	// Init Dataset
	dataset := bqClient.Dataset(cfg.Dataset)
	_, err = dataset.Metadata(ctx)
	tu.FailOnError(t, err)

	// Create Table
	table := dataset.Table(cfg.Table)
	metaData := &bigquery.TableMetadata{
		Schema: *schema,
	}
	err = table.Create(ctx, metaData)
	tu.FailOnError(t, err)
}
