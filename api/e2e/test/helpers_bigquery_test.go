// +build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// Poll BigQuery for log entries every <checkPeriod> until BigQuery returns at least <minCount> rows
// or <timeout> is reached.
//
// If at least <minRowCount> rows is returned, the no of rows retrieved will be returned.
// Else if timeout is reached, error will be non nil.
func waitForLogEntriesInBigQuery(
	table string,
	turingReqID string,
	minRowCount int,
	checkPeriod time.Duration,
	timeout time.Duration,
) (int, error) {

	rowCount := 0

	// Split into project, dataset and table names
	parts := strings.Split(table, ".")
	if len(parts) != 3 {
		return rowCount, fmt.Errorf("unexpected table name format:%s", table)
	}

	// Init BQ Client
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, parts[0])
	if err != nil {
		return rowCount, err
	}

	deadline := time.After(timeout)
	tick := time.Tick(checkPeriod)

	for {
		select {
		case <-deadline:
			return rowCount, fmt.Errorf("timeout reached and rowCount '%d' < minRowCount '%d' in BigQuery",
				rowCount, minRowCount)
		case <-tick:
			// Run query on BigQuery to look for log entries with specific turing_req_id
			q := client.Query(fmt.Sprintf("SELECT turing_req_id FROM `%s` WHERE turing_req_id='%s'",
				table, turingReqID))
			it, err := q.Read(ctx)
			if err != nil {
				return rowCount, err
			}

			rowCount = 0
			for {
				var values []bigquery.Value
				err := it.Next(&values)
				if err == iterator.Done {
					break
				}
				if err != nil {
					return rowCount, err
				}
				rowCount++
			}

			if rowCount >= minRowCount {
				return rowCount, nil
			}
		}
	}
}
