package e2e

import (
	"encoding/json"
	"net/http"

	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"

	"github.com/caraml-dev/turing/api/e2e/new-test/config"
)

var _ = DeployedRouterContext("testdata/create_router_std_ensembler_proprietary_exp.json.tmpl",
	func(routerCtx *RouterContext) {
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
	func(routerCtx *RouterContext) {
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
