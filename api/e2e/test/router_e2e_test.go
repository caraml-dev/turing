package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gavv/httpexpect/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/caraml-dev/turing/api/e2e/test/api"
	"github.com/caraml-dev/turing/api/e2e/test/cluster"
	"github.com/caraml-dev/turing/api/e2e/test/config"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

var _ = DeployedRouterContext("testdata/create_router_nop_logger_proprietary_exp.json.tmpl",
	routerConfig.HTTP, func(routerCtx *RouterContext) {
		var (
			apiE, routerE *httpexpect.Expect
			router        *httpexpect.Object
			version       *httpexpect.Object
		)

		var (
			want, got *httpexpect.Object
		)

		BeforeAll(func() {
			apiE = config.NewHTTPExpect(GinkgoT(), cfg.APIBasePath)
			routerE = config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint)
		})

		AssertPredictResponse := func() {
			got = routerE.POST("/v1/predict").
				WithHeaders(defaultPredictHeaders).
				WithJSON(json.RawMessage(`{"client": {"id": 4}}`)).
				Expect().Status(http.StatusOK).
				JSON().Object()

			AssertResponsePayload(want, got)
		}

		Describe("1.1. Calling router endpoints to fetch predictions", func() {
			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {
					When("valid request", func() {
						It("responds with expected payload", func() {
							want = httpexpect.NewValue(
								GinkgoT(),
								JSONPayload("testdata/responses/router_proprietary_exp_predict.json"),
							).Object()

							AssertPredictResponse()
						})
					})
				})

				Context("POST /v1/batch_predict", func() {
					When("valid request", func() {
						It("responds with an array, that contains individual predictions", func() {
							routerE.POST("/v1/batch_predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`[{"client": {"id": 4}}, {"client": {"id": 7}}]`)).
								Expect().Status(http.StatusOK).
								JSON().Array().
								Equal(JSONPayload("testdata/responses/router_proprietary_exp_batch_predict.json"))
						})
					})
				})
			})
		})

		Describe("1.2. Accessing router logs via Turing API", func() {
			Context("Turing API", func() {
				Context("GET /projects/:projectId/routers/:routerId/logs?component_type=:type", func() {
					for _, componentType := range []string{"router", "ensembler", "enricher"} {
						When(fmt.Sprintf("API is called to fetch %s's logs", componentType), func() {
							It(fmt.Sprintf("responds with %s's log entries", componentType), func() {
								apiE.GET("/projects/{projectId}/routers/{routerId}/logs").
									WithPath("projectId", routerCtx.ProjectID).
									WithPath("routerId", routerCtx.ID).
									WithQuery("component_type", componentType).
									Expect().Status(http.StatusOK).
									JSON().Array().NotEmpty()
							})
						})
					}
				})
			})
		})

		Describe("1.3. Updating router with a new version, that can't be deployed", func() {
			Context("Turing API", func() {
				Context("PUT /projects/:projectId/routers/:routerId", func() {
					When("request valid", func() {
						It("creates a new router's version", func() {
							apiE.PUT("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								WithJSON(JSONPayload("testdata/update_router_high_cpu.json.tmpl", cfg)).
								Expect().Status(http.StatusOK).
								JSON().Object()
						})
					})
				})

				Context("GET /projects/:projectId/routers/:routerId/versions/:version", func() {
					When("new router version can't be deployed", func() {
						It("fails after a timeout", func() {
							Eventually(func(g Gomega) {
								version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 2)

								g.Expect(version.Value("status").Raw()).ShouldNot(Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(Succeed())

							version.
								ValueEqual("status", api.Status.Failed).
								ValueEqual("error", "Requested CPU is more than max permissible")
						})

						It("keeps previously deployed version active", func() {
							router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)
							router.Path("$.config.version").Equal(1)
						})
					})
				})
			})

			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {
					When("valid request", func() {
						It("responds with expected payload", func() {
							Consistently(func(Gomega) {
								AssertPredictResponse()
							}, arbitraryUpdateIntervals...).Should(Succeed())
						})
					})
				})
			})

		})

		Describe("1.4. Undeploying router", func() {
			Context("Turing API", func() {
				When("router is deployed", func() {
					Context("POST /projects/:projectId/routers/:routerId/undeploy", func() {
						It("accepts request and starts undeploying the router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/undeploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)
						})

						It("eventually undeploys the router", func() {
							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(And(
									Not(HaveKey("endpoint")),
									HaveKeyWithValue("status", api.Status.Undeployed),
									HaveKeyWithValue("config", And(
										HaveKeyWithValue("version", BeNumerically("==", 1)),
										HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, defaultDeletionIntervals...).Should(Succeed())
						})
					})
				})
			})
		})

		Describe("1.5. Deploying a version with invalid configuration", func() {
			Context("Turing API", func() {
				Context("POST /projects/:projectId/routers/:routerId/versions/:version/deploy", func() {
					When("version exists", func() {
						It("attempts to deploy this version", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/versions/{version}/deploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								WithPath("version", version.Value("version").Raw()).
								Expect().Status(http.StatusAccepted)

							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)
								// TODO: Why is it pending? - g.Expect(router.Value("status").Raw()).To(Equal(api.Status.Pending))
								g.Expect(router.Value("status").Raw()).To(Equal(api.Status.Undeployed))
							}, arbitraryUpdateIntervals...).Should(Succeed())
						})

						It("fails after a timeout", func() {
							versionID := version.Value("version").Raw()
							Eventually(func(g Gomega) {
								version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, versionID)

								g.Expect(version.Value("status").Raw()).ShouldNot(Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(Succeed())

							version.
								ValueEqual("status", api.Status.Failed).
								ValueEqual("error", "Requested CPU is more than max permissible")
						})

						It("keeps previous valid version as router's current version", func() {
							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(And(
									Not(HaveKey("endpoint")),
									HaveKeyWithValue("status", api.Status.Undeployed),
									HaveKeyWithValue("config", And(
										HaveKeyWithValue("version", BeNumerically("==", 1)),
										HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, arbitraryUpdateIntervals...).Should(Succeed())
						})
					})
				})
			})
		})

		Describe("1.6. Redeploying a router with existing valid configuration", func() {
			Context("Turing API", func() {
				Context("POST /projects/:projectId/routers/:routerId/deploy", func() {
					When("router is not deployed", func() {
						It("attempts to deploy the router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/deploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusAccepted)

							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Value("status").Raw()).To(Equal(api.Status.Pending))
							}, arbitraryUpdateIntervals...).Should(Succeed())
						})

						It("successfully deploys the router", func() {
							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Value("status").Raw()).ShouldNot(Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(Succeed())

							router.
								ValueEqual("status", api.Status.Deployed).
								Path("$.config.version").Equal(1)
						})
					})
				})
			})
		})

		Describe("1.7. Updating router with a new valid configuration", func() {
			Context("Turing API", func() {
				Context("PUT /projects/:projectId/routers/:routerId", func() {
					When("request valid", func() {
						It("creates a new router's version", func() {
							apiE.PUT("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								WithJSON(
									JSONPayload("testdata/update_router_nop_logger_proprietary_exp.json.tmpl", cfg)).
								Expect().Status(http.StatusOK).
								JSON().Object()
						})
					})
				})

				It("successfully deploys a new version", func() {
					Eventually(func(g Gomega) {
						version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 3)

						g.Expect(version.Value("status").Raw()).ShouldNot(Equal(api.Status.Pending))
					}, defaultDeploymentIntervals...).Should(Succeed())

					version.
						ValueEqual("status", api.Status.Deployed).
						NotContainsKey("error")
				})

				It("marks previous version as undeployed", func() {
					Eventually(func(g Gomega) {
						v := api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 1)

						g.Expect(v.Value("status").Raw()).Should(Equal(api.Status.Undeployed))
					}, arbitraryUpdateIntervals...).Should(Succeed())
				})

				It("updates router's configuration to the new version", func() {
					api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID).
						ValueEqual("status", api.Status.Deployed).
						Path("$.config.version").Equal(3)
				})
			})

			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {
					When("when the new version of router is deployed", func() {
						It("responds with expected payload", func() {
							want = httpexpect.NewValue(
								GinkgoT(),
								json.RawMessage(`{"version": "treatment-a"}`),
							).Object()
							AssertPredictResponse()
						})
					})
				})
			})
		})

		Describe("1.8. Deleting router", func() {
			Context("Turing API", func() {
				Context("DELETE /projects/:projectId/routers/:routerId", func() {
					When("router is deployed", func() {
						It("responds with bad request status", func() {
							apiE.DELETE("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusBadRequest).
								JSON().
								Equal(json.RawMessage(`{
                                        "description": "invalid delete request",
                                        "error": "router is currently deployed. Undeploy it first."
                                    }`))
						})
					})
				})

				Context("POST /projects/:projectId/routers/:routerId/undeploy", func() {
					When("router is deployed", func() {
						It("accepts request for undeploying router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/undeploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)
						})

						It("undeploys the router", func() {
							Eventually(func(g Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(And(
									Not(HaveKey("endpoint")),
									HaveKeyWithValue("status", api.Status.Undeployed),
									HaveKeyWithValue("config", And(
										HaveKeyWithValue("version", BeNumerically("==", 3)),
										HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, defaultDeletionIntervals...).Should(Succeed())
						})
					})
				})

				Context("DELETE /projects/:projectId/routers/:routerId", func() {
					When("router is not deployed", func() {
						It("successfully deletes router from the db", func() {
							apiE.DELETE("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)

							apiE.GET("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusNotFound).
								JSON().
								Equal(json.RawMessage(`{
									"description": "router not found",
									"error": "record not found"
								}`))
						})
					})
				})
			})

			Context("Turing Router API", func() {
				Context("POST /v1/predict", func() {
					When("when the router is deleted", func() {
						It("responds with NotFound status code", func() {
							Eventually(func(g Gomega) {
								response := routerE.POST("/v1/predict").
									WithHeaders(defaultPredictHeaders).
									WithJSON(json.RawMessage(`{}`)).
									Expect()

								g.Expect(response.Raw().StatusCode).To(Equal(http.StatusNotFound))
							}, defaultDeletionIntervals...).Should(Succeed())
						})
					})
				})
			})

			Context("K8s API", func() {
				When("the router is deleted", func() {
					It("should remove all its k8s resources from the cluster", func() {
						routerName := router.Value("name").String().Raw()
						Eventually(func(g Gomega) {
							k8sResources, err := cluster.ListRouterResources(cfg.Project.Name, routerName)

							g.Expect(err).ShouldNot(HaveOccurred())
							g.Expect(k8sResources).Should(And(
								HaveField("KnativeServices", BeEmpty()),
								HaveField("K8sServices", BeEmpty()),
								HaveField("IstioServices", BeEmpty()),
								HaveField("K8sDeployments", BeEmpty()),
								HaveField("ConfigMaps", BeEmpty()),
								HaveField("Secrets", BeEmpty()),
								HaveField("PVCs", BeEmpty()),
							))
						}, defaultDeletionIntervals...).Should(Succeed())
					})
				})
			})
		})
	})
