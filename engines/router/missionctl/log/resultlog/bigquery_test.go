package resultlog

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"

	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
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

func TestBigQueryLoggerGetData(t *testing.T) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := io.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)

	// Create new logger
	testLogger := &bigQueryLogger{}

	// Create a TuringResultLogEntry record and add the data
	timestamp := time.Date(2000, 2, 1, 4, 5, 6, 7, time.UTC)
	entry := NewTuringResultLog(turingReqID, timestamp, req.Header, string(reqBody))
	entry.RouterVersion = "turing-router-1"
	AddResponse(entry, "experiment", `{"key": "experiment_data"}`, nil, "")
	AddResponse(entry, "enricher", `{"key": "enricher_data"}`, nil, "")
	AddResponse(entry, "router", `{"key": "router_data"}`, nil, "")
	AddResponse(entry, "ensembler", "", nil, "Error Response")

	// Get the log data and validate
	logData := testLogger.getLogData(entry)
	// Cast to map[string]bigquery.Value
	if logMap, ok := logData.(map[string]bigquery.Value); ok {
		// Turing Request ID
		assert.Equal(t, turingReqID, logMap["turing_req_id"])

		// Router Version
		assert.Equal(t, "turing-router-1", logMap["router_version"])

		// Timestamp
		assert.Equal(t, "2000-02-01T04:05:06.000000007Z", logMap["event_timestamp"])

		// Request
		if requestData, ok := logMap["request"].(map[string]interface{}); ok {
			assert.Equal(t, []map[string]interface{}{
				{
					"key":   "Req_id",
					"value": "test_req_id",
				},
			}, requestData["header"])
			assert.Equal(t, `{"customer_id": "test_customer"}`, requestData["body"])
		} else {
			tu.FailOnError(t, fmt.Errorf("Cannot cast request log to expected type"))
		}

		// Experiment
		if respObj, ok := logMap["experiment"].(map[string]interface{}); ok {
			assert.Equal(t, `{"key": "experiment_data"}`, respObj["response"])
			assert.Equal(t, nil, respObj["error"])
		} else {
			tu.FailOnError(t, fmt.Errorf("Cannot cast experiment log to expected type"))
		}

		// Enricher
		if respObj, ok := logMap["enricher"].(map[string]interface{}); ok {
			assert.Equal(t, `{"key": "enricher_data"}`, respObj["response"])
			assert.Equal(t, nil, respObj["error"])
		} else {
			tu.FailOnError(t, fmt.Errorf("Cannot cast enricher log to expected type"))
		}

		// Router
		if respObj, ok := logMap["router"].(map[string]interface{}); ok {
			assert.Equal(t, `{"key": "router_data"}`, respObj["response"])
			assert.Equal(t, nil, respObj["error"])
		} else {
			tu.FailOnError(t, fmt.Errorf("Cannot cast router log to expected type"))
		}

		// Ensembler
		if requestObj, ok := logMap["ensembler"].(map[string]interface{}); ok {
			assert.Equal(t, nil, requestObj["body"])
			assert.Equal(t, "Error Response", requestObj["error"])
		} else {
			tu.FailOnError(t, fmt.Errorf("Cannot cast ensembler log to expected type"))
		}
	} else {
		tu.FailOnError(t, fmt.Errorf("Cannot cast log result to expected type"))
	}
}
