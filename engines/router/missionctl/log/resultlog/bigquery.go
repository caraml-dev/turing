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

	"cloud.google.com/go/bigquery"
	"go.einride.tech/protobuf-bigquery/encoding/protobq"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
)

// bqLogEntry wraps a TuringResultLogEntry and implements the bigquery.ValueSaver interface
type bqLogEntry struct {
	resultLog *turing.TuringResultLogMessage
}

// Save implements the ValueSaver interface on bqLogEntry, for saving the data to BigQuery
func (e *bqLogEntry) Save() (map[string]bigquery.Value, string, error) {
	var kvPairs map[string]bigquery.Value
	bytes, err := protoJSONMarshaller.Marshal(e.resultLog)
	if err != nil {
		return kvPairs, "", err
	}
	// Unmarshal into map[string]bigquery.Value
	err = json.Unmarshal(bytes, &kvPairs)
	if err != nil {
		return kvPairs, "", errors.Wrapf(err, "Error unmarshaling the result log for save to BQ")
	}

	// Special handling: Update request, experiment, enricher, router and ensembler headers to a list of records,
	// expected by BQ.
	// It seems protobq.Marshal will be adding support for map[string]string that would help simplify the
	// implementation of Save().
	kvPairs["request"] = bigquery.Value(map[string]interface{}{
		"header": formatBQLogEntryHeader(e.resultLog.Request.Header),
		"body":   e.resultLog.Request.Body,
	})
	kvPairs["experiment"] = formatBQLogEntryResponse(e.resultLog.Experiment)
	kvPairs["enricher"] = formatBQLogEntryResponse(e.resultLog.Enricher)
	kvPairs["router"] = formatBQLogEntryResponse(e.resultLog.Router)
	kvPairs["ensembler"] = formatBQLogEntryResponse(e.resultLog.Ensembler)

	return kvPairs, "", nil
}

// formatBQLogEntryResponse formats the entire response manually due to the manual handling required for the headers
func formatBQLogEntryResponse(response *turing.Response) bigquery.Value {
	if response == nil {
		return nil
	}
	bgLogEntryComponents := map[string]interface{}{}

	if response.Header != nil {
		bgLogEntryComponents["header"] = formatBQLogEntryHeader(response.Header)
	}

	if response.Response != "" {
		bgLogEntryComponents["response"] = response.Response
	}

	if response.Error != "" {
		bgLogEntryComponents["error"] = response.Error
	}
	return bigquery.Value(bgLogEntryComponents)
}

// formatBQLogEntryHeader formats header values in a map into a list of header values
func formatBQLogEntryHeader(headerMap map[string]string) []map[string]interface{} {
	headers := []map[string]interface{}{}
	for key, value := range headerMap {
		headers = append(headers, map[string]interface{}{
			"key":   key,
			"value": value,
		})
	}
	return headers
}

// BigQueryLogger extends the TuringResultLogger interface and defines additional
// methods on the logger
type BigQueryLogger interface {
	TuringResultLogger
	getLogData(message *turing.TuringResultLogMessage) interface{}
}

// bigQueryLogger implements the BigQueryLogger interface and wraps the bigquery.Client
// and other necessary information to save the data to BigQuery
type bigQueryLogger struct {
	dataset  string
	table    string
	bqClient *bigquery.Client
	schema   *bigquery.Schema
}

// NewBigQueryLogger creates a new BigQueryLogger
func NewBigQueryLogger(cfg *config.BQConfig) (BigQueryLogger, error) {
	ctx := context.Background()
	bqClient, err := bigquery.NewClient(ctx, cfg.Project)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to initialize BigQuery Client")
	}
	// Create the BigQuery logger
	bqLogger := &bigQueryLogger{
		dataset:  cfg.Dataset,
		table:    cfg.Table,
		bqClient: bqClient,
		schema:   getTuringResultTableSchema(),
	}
	// Set up Turing Result table
	err = bqLogger.setUpTuringTable()
	if err != nil {
		return nil, err
	}
	return bqLogger, nil
}

// write satisfies the TuringResultLogger interface
func (l *bigQueryLogger) write(t *turing.TuringResultLogMessage) error {
	// Create an inserter
	ins := l.bqClient.Dataset(l.dataset).Table(l.table).Inserter()

	// Each request has a 10MB limit, and each record has a 1MB limit.
	// Currently inserting one at a time.
	items := []*bqLogEntry{{t}}

	// Write the data and return any errors
	if err := ins.Put(context.Background(), items); err != nil {
		return errors.Wrapf(err, "Error during streaming insert")
	}
	return nil
}

// getLogData returns the log information as a generic interface{} object. Internally, it calls
// the Save method defined on the bqLogEntry structure which implements the
// bigquery.ValueSaver interface and returns the log data as a map. This can be returned
// as is for logging by other loggers whose destination is a BQ table.
func (l *bigQueryLogger) getLogData(turLogEntry *turing.TuringResultLogMessage) interface{} {
	entry := &bqLogEntry{turLogEntry}
	record, _, err := entry.Save()
	if err != nil {
		log.Glob().Warnf("failed to create log entry %s", err)
	}
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
	metadata, err := table.Metadata(ctx)

	// If not, create
	if err != nil {
		err = createTuringResultTable(&ctx, table, l.schema)
		if err != nil {
			return errors.Wrapf(err, "Failed creating BigQuery table %s", l.table)
		}
	} else {
		// Table exists, compare schema
		schema, isUpdated, err := compareTableSchema(&metadata.Schema, l.schema)
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
			Field:                  "event_timestamp",
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
func getTuringResultTableSchema() *bigquery.Schema {
	schema := protobq.InferSchema(&turing.TuringResultLogMessage{})
	return &schema
}
