package e2e

import (
	"context"
	"encoding/json"
	"net/http"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gavv/httpexpect/v2"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/caraml-dev/turing/api/e2e/test/config"
	"github.com/caraml-dev/turing/api/e2e/test/matcher"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

var _ = DeployedRouterContext("testdata/create_router_std_ensembler_proprietary_exp.json.tmpl",
	routerConfig.HTTP, func(routerCtx *RouterContext) {
		var routerE *httpexpect.Expect

		ginkgo.BeforeAll(func() {
			routerE = config.NewHTTPExpect(ginkgo.GinkgoT(), routerCtx.Endpoint)
		})

		ginkgo.Describe("Calling router endpoints to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {

					ginkgo.When("called with {client: {id: 4}} in the payload", func() {
						ginkgo.It("responds with data from the `control` route", func() {
							routerE.POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"client": {"id": 4}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().IsEqual(json.RawMessage(`{"version": "control"}`))
						})
					})

					ginkgo.When("called with {client: {id: 7}} in the payload", func() {
						ginkgo.It("responds with data from the `treatment-a` route", func() {
							routerE.POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"client": {"id": 7}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().IsEqual(json.RawMessage(`{"version": "treatment-a"}`))
						})
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_with_traffic_rules.json.tmpl",
	routerConfig.HTTP, func(routerCtx *RouterContext) {
		ginkgo.Describe("Calling router endpoints to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {
					var want, got *httpexpect.Object

					ginkgo.AfterEach(func() {
						AssertResponsePayload(want, got)
					})

					ginkgo.When("request satisfies the first traffic rule", func() {
						ginkgo.It("responds with responses from `treatment-a` route", func() {
							want = httpexpect.
								NewValue(ginkgo.GinkgoT(),
									JSONPayload("testdata/responses/traffic_rules/traffic-rule-1.json")).
								Object()

							got = config.NewHTTPExpect(ginkgo.GinkgoT(), routerCtx.Endpoint).
								POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithHeader("X-Region", "region-a").
								WithJSON(json.RawMessage(`{}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Object()
						})
					})

					ginkgo.When("request satisfies the second traffic rule", func() {
						ginkgo.It("responds with responses from `treatment-b` route", func() {
							want = httpexpect.
								NewValue(ginkgo.GinkgoT(),
									JSONPayload("testdata/responses/traffic_rules/traffic-rule-2.json")).
								Object()

							got = config.NewHTTPExpect(ginkgo.GinkgoT(), routerCtx.Endpoint).
								POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"service_type": {"id": "service-type-b"}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Object()
						})
					})

					ginkgo.When("request satisfies no traffic rules", func() {
						ginkgo.It("responds with responses from `control` route", func() {
							want = httpexpect.
								NewValue(ginkgo.GinkgoT(), JSONPayload("testdata/responses/traffic_rules/no-rules.json")).
								Object()

							got = config.NewHTTPExpect(ginkgo.GinkgoT(), routerCtx.Endpoint).
								POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"service_type": {"id": "service-type-c"}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Object()
						})
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_upi_simple.json.tmpl", routerConfig.UPI,
	func(routerCtx *RouterContext) {
		ginkgo.Describe("Calling UPI router endpoint to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.When("send UPI PredictValues", func() {
					ginkgo.It("responds successfully", func() {
						conn, _ := grpc.Dial(routerCtx.Endpoint,
							grpc.WithTransportCredentials(insecure.NewCredentials()))
						defer conn.Close()

						client := upiv1.NewUniversalPredictionServiceClient(conn)
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
									Name:        "country_code",
									Type:        upiv1.Type_TYPE_STRING,
									StringValue: "ID",
								},
								{
									Name:        "order_id",
									Type:        upiv1.Type_TYPE_STRING,
									StringValue: "12345",
								},
							},
						}
						headers := metadata.New(map[string]string{"region": "region-a"})
						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						gomega.Expect(err).To(gomega.BeNil())
						gomega.Expect(resp.Metadata.Models[0].Name, err).To(gomega.Equal("control"))
						gomega.Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_upi_with_std_ensembler.json.tmpl", routerConfig.UPI,
	func(routerCtx *RouterContext) {
		ginkgo.Describe("Calling UPI router endpoint to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.When("send UPI PredictValues that generate treatment-a", func() {
					ginkgo.It("responds successfully using treatment-a", func() {
						conn, _ := grpc.Dial(routerCtx.Endpoint,
							grpc.WithTransportCredentials(insecure.NewCredentials()))
						defer conn.Close()

						client := upiv1.NewUniversalPredictionServiceClient(conn)
						headers := metadata.New(map[string]string{"region": "region-a"})
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
									StringValue: "7",
								},
							},
						}

						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						gomega.Expect(err).To(gomega.BeNil())
						gomega.Expect(resp.Metadata.Models[0].Name, err).To(gomega.Equal("treatment-a"))
						gomega.Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_upi_with_traffic_rules.json.tmpl", routerConfig.UPI,
	func(routerCtx *RouterContext) {
		ginkgo.Describe("Calling UPI router endpoint to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.When("send UPI PredictValues that satisfy traffic rule rule-1", func() {
					ginkgo.It("responds successfully using treatment-a endpoint", func() {
						conn, _ := grpc.Dial(routerCtx.Endpoint,
							grpc.WithTransportCredentials(insecure.NewCredentials()))
						defer conn.Close()

						client := upiv1.NewUniversalPredictionServiceClient(conn)
						headers := metadata.New(map[string]string{})
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
									StringValue: "7",
								},
							},
						}

						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						gomega.Expect(err).To(gomega.BeNil())
						gomega.Expect(resp.Metadata.Models[0].Name, err).To(gomega.Equal("treatment-a"))
						gomega.Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
				ginkgo.When("sent UPI PredictValues doesn't satisfy any traffic rule", func() {
					ginkgo.It("responds successfully using control endpoint", func() {
						conn, _ := grpc.Dial(routerCtx.Endpoint,
							grpc.WithTransportCredentials(insecure.NewCredentials()))
						defer conn.Close()

						client := upiv1.NewUniversalPredictionServiceClient(conn)
						headers := metadata.New(map[string]string{})
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

						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						gomega.Expect(err).To(gomega.BeNil())
						gomega.Expect(resp.Metadata.Models[0].Name, err).To(gomega.Equal("control"))
						gomega.Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})
