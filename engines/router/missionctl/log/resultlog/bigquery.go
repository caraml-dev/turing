package resultlog

/*
This file defines the BigQuery result logger. There are 2 parts:
1. The bigQueryLogger.write() which satisfies the TuringResultLogger interface defined in
   missionctl/log/resultlog/resultlog.go, saving the log entry to BigQuery using the streaming
   insert API, and the bigQueryLogger.getLogData() which is used by the fluentd logger -
   this is the part used for logging.
2. bigQueryLogger.setUpTuringTable() and supporting methods to check that the required dataset
   exists and the table exists (if so, validate the schema and if not, create the table).
*/

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/gojek/turing/engines/router/missionctl/config"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

// bqResponseKeys captures the response keys that are applicable to the defined table schema.
// Keys not belonging to this list will be dropped when saving the log entry to BQ.
var bqResponseKeys = []string{
	ResultLogKeys.Experiment,
	ResultLogKeys.Enricher,
	ResultLogKeys.Router,
	ResultLogKeys.Ensembler,
}

// bqLogEntry wraps a TuringResultLogEntry and implements the bigquery.ValueSaver interface
type bqLogEntry struct {
	appName string
	*TuringResultLogEntry
}

// newBqLogEntry creates a new bqLogEntry from the given TuringResultLogEntry
func newBqLogEntry(appName string, turLogEntry *TuringResultLogEntry) *bqLogEntry {
	return &bqLogEntry{
		appName,
		turLogEntry,
	}
}

// Save implements the ValueSaver interface on bqLogEntry, for saving the data to BigQuery
func (e *bqLogEntry) Save() (map[string]bigquery.Value, string, error) {
	// Get Turing Request Id
	turingReqID, err := turingctx.GetRequestID(*e.ctx)
	if err != nil {
		return map[string]bigquery.Value{}, "", err
	}

	// Convert the request header to json
	requestHeader, err := json.Marshal(e.request.Header)
	if err != nil {
		return map[string]bigquery.Value{}, "", err
	}

	// Create the record for saving the data to BQ
	record := map[string]bigquery.Value{
		"turing_req_id":  turingReqID,
		"ts":             e.timestamp.Format(time.RFC3339Nano),
		"router_version": e.appName,
		"request": map[string]string{
			"header": string(requestHeader),
			"body":   string(e.request.Body),
		},
	}
	// Add optional fields defined in bqResponseKeys, if they exist
	for _, key := range bqResponseKeys {
		if v, exist := e.responses[key]; exist {
			record[key] = map[string]string{
				"response": string(v.Response),
				"error":    v.Error,
			}
		}
	}

	return record, "", nil
}

// BigQueryLogger extends the TuringResultLogger interface and defines additional
// methods on the logger
type BigQueryLogger interface {
	TuringResultLogger
	getLogData(*TuringResultLogEntry) interface{}
}

// bigQueryLogger implements the BigQueryLogger interface and wraps the bigquery.Client
// and other necessary information to save the data to BigQuery
type bigQueryLogger struct {
	appName  string
	dataset  string
	table    string
	bqClient *bigquery.Client
}

// newBigQueryLogger creates a new BigQueryLogger
func newBigQueryLogger(appName string, cfg *config.BQConfig) (BigQueryLogger, error) {
	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, cfg.Project)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to initialize BigQuery Client")
	}
	// Create the BigQuery logger
	bqLogger := &bigQueryLogger{
		appName:  appName,
		dataset:  cfg.Dataset,
		table:    cfg.Table,
		bqClient: bqClient,
	}
	// Set up Turing Result table
	err = bqLogger.setUpTuringTable()
	if err != nil {
		return nil, err
	}
	return bqLogger, nil
}

// write satisfies the TuringResultLogger interface
func (l *bigQueryLogger) write(t *TuringResultLogEntry) error {
	entry := newBqLogEntry(l.appName, t)
	// Create an inserter
	ins := l.bqClient.Dataset(l.dataset).Table(l.table).Inserter()
	// Each request has a 10MB limit, and each record has a 1MB limit.
	// Insert one at a time.
	items := []*bqLogEntry{
		entry,
	}
	// Write the data and return any errors
	if err := ins.Put(context.Background(), items); err != nil {
		return errors.Wrapf(err, "Error during streaming insert")
	}
	return nil
}

// getLogData returns the log information as a generic interface{} object. Internally, it calls
// the Save method defined on the bqLogEntry structure which implements the
// bigquery.ValueSaver interface and returns the log data as a map. This can be returned
// as is for logging by other loggers.
func (l *bigQueryLogger) getLogData(turLogEntry *TuringResultLogEntry) interface{} {
	entry := newBqLogEntry(l.appName, turLogEntry)
	record, _, _ := entry.Save()
	return record
}

// setUpTuringTable checks that the logging table is set up in BQ as expected.
// If the specified dataset does not exist in the project, it returns an error.
// If the dataset + table exists and the schema does not match the expected,
// an error is returned as well. If the dataset exists but not the table, a
// new table is created.
func (l *bigQueryLogger) setUpTuringTable() error {
	ctx := context.Background()

	// Check that the dataset exists
	dataset := l.bqClient.Dataset(l.dataset)
	_, err := dataset.Metadata(ctx)
	if err != nil {
		return errors.Wrapf(err, "BigQuery dataset %s not found", l.dataset)
	}

	// Check if the table exists
	table := dataset.Table(l.table)
	schema := getTuringResultTableSchema()
	metadata, err := table.Metadata(ctx)

	// If not, create
	if err != nil {
		err = createTuringResultTable(&ctx, table, &schema)
		if err != nil {
			return errors.Wrapf(err, "Failed creating BigQuery table %s", l.table)
		}
	} else {
		// Table exists, compare schema
		schema, isUpdated, err := compareTableSchema(&metadata.Schema, &schema)
		if err != nil {
			return errors.Wrapf(err, "Unexpected schema for BigQuery table %s", l.table)
		}
		// Update schema, if it changed
		if isUpdated {
			update := bigquery.TableMetadataToUpdate{
				Schema: *schema,
			}
			if _, err := table.Update(ctx, update, metadata.ETag); err != nil {
				return err
			}
		} else {
			// No update to schema required, check that we have the required perms
			// for data write
			return checkBQTableWritePermissions(table)
		}
	}

	return nil
}

// checkPermissions checks that the BQ client has the required permissions on the dataset
func checkBQTableWritePermissions(table *bigquery.Table) error {
	// Ref: https://cloud.google.com/bigquery/docs/access-control
	requiredPerms := []string{
		"bigquery.tables.get",
		"bigquery.tables.getData",
		"bigquery.tables.update",
	}
	perms, err := table.IAM().TestPermissions(context.Background(), requiredPerms)
	if err != nil {
		return errors.Newf(errors.BadConfig,
			"Error checking IAM permissions on the BQ table: %s", err.Error())
	}
	if len(perms) < len(requiredPerms) {
		return errors.Newf(errors.BadConfig,
			"Insufficient permissions. Got: %s; Want: %s",
			strings.Join(perms, ","),
			strings.Join(requiredPerms, ","),
		)
	}
	return nil
}

// createTuringResultTable creates the specified table if not exists
func createTuringResultTable(
	ctx *context.Context,
	table *bigquery.Table,
	schema *bigquery.Schema,
) error {
	// Set partitioning
	metaData := &bigquery.TableMetadata{
		Schema: *schema,
		TimePartitioning: &bigquery.TimePartitioning{
			Field:                  "ts",
			RequirePartitionFilter: false,
		},
	}

	// Create the table
	if err := table.Create(*ctx, metaData); err != nil {
		return err
	}

	return nil
}

// compareTableSchema validates the important properties of each field in the schema
// recursively. If the expected schema has more columns than the actual, and these
// columns are nullable, they are added to the actual schema and the 'updated' flag
// is set to true. The (updated) actual schema, the flag and any error is returned.
func compareTableSchema(
	tableSchema *bigquery.Schema,
	expectedSchema *bigquery.Schema,
) (*bigquery.Schema, bool, error) {
	isUpdated := false

	// Create a map of the tableSchema column name to the field
	tableSchemaMap := map[string]*bigquery.FieldSchema{}
	for _, item := range *tableSchema {
		tableSchemaMap[item.Name] = item
	}

	// For each field in the expected schema, add it to the tableSchema if it
	// doesn't already exist. Compare the properties otherwise.
	for _, ef := range *expectedSchema {
		var af *bigquery.FieldSchema
		var ok bool
		if af, ok = tableSchemaMap[ef.Name]; ok {
			// Compare current schema
			if af.Name != ef.Name ||
				af.Type != ef.Type ||
				af.Required != ef.Required ||
				af.Repeated != ef.Repeated {
				return tableSchema, false, errors.Newf(errors.BadConfig,
					"BigQuery schema mismatch for field %s", ef.Name)
			}
			// Compare nested schema
			nestedSchema, itemUpdated, err := compareTableSchema(&af.Schema, &ef.Schema)
			// If error, return
			if err != nil {
				return tableSchema, false, err
			}
			// Save the (new) nested schema to the current field
			af.Schema = *nestedSchema
			// Set the overall updated flag
			isUpdated = isUpdated || itemUpdated
		} else if !ef.Required {
			// Append NULLABLE missing field to the tableSchema
			*tableSchema = append(*tableSchema, ef)
			isUpdated = true
		} else {
			// Return error
			return tableSchema, false, errors.Newf(errors.BadConfig,
				"Cannot add Required field %s to the existing BQ table", ef.Name)
		}
	}

	// At the end of any additions, the two schemas must have the same number of fields
	if len(*tableSchema) != len(*expectedSchema) {
		return tableSchema, false, errors.Newf(errors.BadConfig, "BigQuery schema mismatch")
	}

	return tableSchema, isUpdated, nil
}

// getTuringResultTableSchema returns the expected schema defined for logging results
// to BigQuery
func getTuringResultTableSchema() bigquery.Schema {
	schema := bigquery.Schema{
		{Name: "turing_req_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "ts", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "router_version", Type: bigquery.StringFieldType, Required: false},
		{Name: "request", Type: bigquery.RecordFieldType,
			Required: true,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "header", Type: bigquery.StringFieldType},
				{Name: "body", Type: bigquery.StringFieldType},
			},
		},
		{Name: "experiment", Type: bigquery.RecordFieldType,
			Required: false,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "response", Type: bigquery.StringFieldType},
				{Name: "error", Type: bigquery.StringFieldType},
			},
		},
		{Name: "enricher", Type: bigquery.RecordFieldType,
			Required: false,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "response", Type: bigquery.StringFieldType},
				{Name: "error", Type: bigquery.StringFieldType},
			},
		},
		{Name: "router", Type: bigquery.RecordFieldType,
			Required: false,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "response", Type: bigquery.StringFieldType},
				{Name: "error", Type: bigquery.StringFieldType},
			},
		},
		{Name: "ensembler", Type: bigquery.RecordFieldType,
			Required: false,
			Repeated: false,
			Schema: bigquery.Schema{
				{Name: "response", Type: bigquery.StringFieldType},
				{Name: "error", Type: bigquery.StringFieldType},
			},
		},
	}
	return schema
}
