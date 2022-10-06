//go:build e2e

package e2e

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/caraml-dev/turing/api/turing/models"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*
Steps:
Create a new router with valid config for the router.
a. Test GET router immediately > empty config
b. Wait for success response from deployment
c. Test GET router version > status shows "deployed"
d. Test GET router > config section shows version 1, status "deployed"
e. Test cluster that deployments exist
f. Make a request to the router, validate the response.
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
			conn, err := grpc.Dial(router.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			assert.NoError(t, err)
			defer conn.Close()

			c := upiv1.NewUniversalPredictionServiceClient(conn)
			r, err := c.PredictValues(context.Background(), &upiv1.PredictValuesRequest{})
			assert.NoError(t, err)
			t.Log(r.String())

			endpoint2 := router.Endpoint[:len(router.Endpoint)-3]
			t.Log("Testing router endpoint2: " + endpoint2)
			conn, err := grpc.Dial(endpoint2, grpc.WithTransportCredentials(insecure.NewCredentials()))
			assert.NoError(t, err)
			defer conn.Close()

			c := upiv1.NewUniversalPredictionServiceClient(conn)
			r, err := c.PredictValues(context.Background(), &upiv1.PredictValuesRequest{})
			assert.NoError(t, err)
			t.Log(r.String())
		},
		nil,
	)
}
