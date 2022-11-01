package e2e

import (
	"context"
	"encoding/json"
	"net/http"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		BeforeAll(func() {
			routerE = config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint)
		})

		Describe("Calling router endpoints to fetch predictions", func() {
			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {

					When("called with {client: {id: 4}} in the payload", func() {
						It("responds with data from the `treatment-a` route", func() {
							routerE.POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"client": {"id": 4}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Equal(json.RawMessage(`{"version": "treatment-a"}`))
						})
					})

					When("called with {client: {id: 7}} in the payload", func() {
						It("responds with data from the `control` route", func() {
							routerE.POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"client": {"id": 7}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Equal(json.RawMessage(`{"version": "control"}`))
						})
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_with_traffic_rules.json.tmpl",
	routerConfig.HTTP, func(routerCtx *RouterContext) {
		Describe("Calling router endpoints to fetch predictions", func() {
			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {
					var want, got *httpexpect.Object

					AfterEach(func() {
						AssertResponsePayload(want, got)
					})

					When("request satisfies the first traffic rule", func() {
						It("responds with responses from `control` and `treatment-a` routes", func() {
							want = httpexpect.
								NewValue(GinkgoT(), JSONPayload("testdata/responses/traffic_rules/traffic-rule-1.json")).
								Object()

							got = config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint).
								POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithHeader("X-Region", "region-a").
								WithJSON(json.RawMessage(`{}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Object()
						})
					})

					When("request satisfies the second traffic rule", func() {
						It("responds with responses from `control` and `treatment-b` routes", func() {
							want = httpexpect.
								NewValue(GinkgoT(), JSONPayload("testdata/responses/traffic_rules/traffic-rule-2.json")).
								Object()

							got = config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint).
								POST("/v1/predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`{"service_type": {"id": "service-type-b"}}`)).
								Expect().
								Status(http.StatusOK).
								JSON().Object()
						})
					})

					When("request satisfies no traffic rules", func() {
						It("responds with responses from `control` route only", func() {
							want = httpexpect.
								NewValue(GinkgoT(), JSONPayload("testdata/responses/traffic_rules/no-rules.json")).
								Object()

							got = config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint).
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
		Describe("Calling UPI router endpoint to fetch predictions", func() {
			Context("Turing Router API", func() {
				When("send UPI PredictValues", func() {
					It("responds successfully", func() {
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

						Expect(err).To(BeNil())
						Expect(resp.Metadata.Models[0].Name, err).To(Equal("control"))
						Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_upi_with_std_ensembler.json.tmpl", routerConfig.UPI,
	func(routerCtx *RouterContext) {
		Describe("Calling UPI router endpoint to fetch predictions", func() {
			Context("Turing Router API", func() {
				When("send UPI PredictValues that generate treatment-a", func() {
					It("responds successfully using treatment-a", func() {
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
									StringValue: "4",
								},
							},
						}

						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						Expect(err).To(BeNil())
						Expect(resp.Metadata.Models[0].Name, err).To(Equal("treatment-a"))
						Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})

var _ = DeployedRouterContext("testdata/create_router_upi_with_traffic_rules.json.tmpl", routerConfig.UPI,
	func(routerCtx *RouterContext) {
		Describe("Calling UPI router endpoint to fetch predictions", func() {
			Context("Turing Router API", func() {
				When("send UPI PredictValues that satisfy traffic rule rule-1", func() {
					It("responds successfully using treatment-a endpoint", func() {
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
									StringValue: "1",
								},
							},
						}

						resp, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
							upiRequest)

						Expect(err).To(BeNil())
						Expect(resp.Metadata.Models[0].Name, err).To(Equal("treatment-a"))
						Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
				When("sent UPI PredictValues doesn't satisfy any traffic rule", func() {
					It("responds successfully using control endpoint", func() {
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

						Expect(err).To(BeNil())
						Expect(resp.Metadata.Models[0].Name, err).To(Equal("control"))
						Expect(resp.PredictionResultTable).To(matcher.ProtoEqual(upiRequest.PredictionTable))
					})
				})
			})
		})
	})
