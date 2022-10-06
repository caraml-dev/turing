//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
Create a new upi router with valid config for the router.
No traffic rules, enricher, ensembler or experiment engine.
Use UPI Client to call router directly.
*/
func TestUpiRouter(t *testing.T) {
	// Create router
	t.Log("Creating router")
	data := makeRouterPayload(
		filepath.Join("testdata", "create_router_upi_simple.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Log("Testing router endpoint: " + router.Endpoint)
			t.Log("Route Endpoint: " + globalTestContext.MockUpiServerEndpoint)
			expectedEndpoint := fmt.Sprintf(
				"%s-turing-router.%s.%s:80",
				router.Name,
				globalTestContext.ProjectName,
				globalTestContext.KServiceDomain,
			)
			assert.Equal(t, expectedEndpoint, router.Endpoint)
			conn, err := grpc.Dial(router.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			assert.NoError(t, err)
			defer conn.Close()

			c := upiv1.NewUniversalPredictionServiceClient(conn)
			upiRequest := &upiv1.PredictValuesRequest{
				PredictionTable: &upiv1.Table{
					Name: "Test",
					Columns: []*upiv1.Column{
						{
							Name: "col1",
							Type: upiv1.Type_TYPE_DOUBLE,
						},
					},
					Rows: []*upiv1.Row{
						{
							RowId: "1",
							Values: []*upiv1.Value{
								{},
							},
						},
					},
				},
			}
			r, err := c.PredictValues(context.Background(), upiRequest)
			assert.NoError(t, err)
			// Upi echo server will send request table in result table and metadata, test to check marshaling is not erroneous
			assert.Equal(t, upiRequest.GetPredictionTable(), r.GetPredictionResultTable())
			assert.NotNil(t, r.GetMetadata())
			t.Log(r.String())
		},
		nil,
	)
}
