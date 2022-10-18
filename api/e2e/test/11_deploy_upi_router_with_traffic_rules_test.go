//go:build e2e

package e2e

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestDeployUPIRouterWithTrafficRules(t *testing.T) {
	// Create router
	t.Log("Creating router")
	data := makeRouterPayload(
		filepath.Join("testdata", "create_router_upi_with_traffic_rules.json.tmpl"),
		globalTestContext)

	withDeployedRouter(t, data,
		func(router *models.Router) {
			t.Logf("Testing router endpoint: %s", router.Endpoint)
			t.Logf("Control Endpoint: %s", globalTestContext.MockserverUPIControlEndpoint)
			t.Logf("Treatment A Endpoint: %s", globalTestContext.MockserverUPIAEndpoint)
			t.Logf("Treatment B Endpoint: %s", globalTestContext.MockserverUPIBEndpoint)

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

			client := upiv1.NewUniversalPredictionServiceClient(conn)
			t.Log("Testing router endpoint with request that satisfy traffic rule rule-1")
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
				PredictionContext: []*upiv1.Variable{
					{
						Name:        "client_id",
						Type:        upiv1.Type_TYPE_STRING,
						StringValue: "1",
					},
				},
			}

			headers := metadata.New(map[string]string{})
			withUPIRouterResponse(t, client, headers, upiRequest, func(response *upiv1.PredictValuesResponse) {
				assert.Equal(t, "treatment-a", response.Metadata.Models[0].Name)
				assert.True(t, proto.Equal(upiRequest.PredictionTable, response.PredictionResultTable))
			})

			t.Log("Testing router endpoint with the request, that doesn't satisfy any traffic rule")
			headers = metadata.New(map[string]string{})
			upiRequest = &upiv1.PredictValuesRequest{
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

			withUPIRouterResponse(t, client, headers, upiRequest, func(response *upiv1.PredictValuesResponse) {
				assert.Equal(t, "control", response.Metadata.Models[0].Name)
				assert.True(t, proto.Equal(upiRequest.PredictionTable, response.PredictionResultTable))
			})
		},
		func(router *models.Router) {
			deleteExperiments(
				globalTestContext.clusterClients,
				globalTestContext.ProjectName,
				[]struct {
					Name       string
					MaxVersion int
				}{
					{
						Name:       router.Name,
						MaxVersion: 1,
					},
				},
			)
		})
}
