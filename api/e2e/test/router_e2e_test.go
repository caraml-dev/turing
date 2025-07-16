package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gavv/httpexpect/v2"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

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

		ginkgo.BeforeAll(func() {
			apiE = config.NewHTTPExpect(ginkgo.GinkgoT(), cfg.APIBasePath)
			routerE = config.NewHTTPExpect(ginkgo.GinkgoT(), routerCtx.Endpoint)
		})

		AssertPredictResponse := func() {
			got = routerE.POST("/v1/predict").
				WithHeaders(defaultPredictHeaders).
				WithJSON(json.RawMessage(`{"client": {"id": 4}}`)).
				Expect().Status(http.StatusOK).
				JSON().Object()

			AssertResponsePayload(want, got)
		}

		ginkgo.Describe("1.1. Calling router endpoints to fetch predictions", func() {
			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {
					ginkgo.When("valid request", func() {
						ginkgo.It("responds with expected payload", func() {
							want = httpexpect.NewValue(
								ginkgo.GinkgoT(),
								JSONPayload("testdata/responses/router_proprietary_exp_predict.json"),
							).Object()

							AssertPredictResponse()
						})
					})
				})

				ginkgo.Context("POST /v1/batch_predict", func() {
					ginkgo.When("valid request", func() {
						ginkgo.It("responds with an array, that contains individual predictions", func() {
							routerE.POST("/v1/batch_predict").
								WithHeaders(defaultPredictHeaders).
								WithJSON(json.RawMessage(`[{"client": {"id": 4}}, {"client": {"id": 7}}]`)).
								Expect().Status(http.StatusOK).
								JSON().Array().
								IsEqual(JSONPayload("testdata/responses/router_proprietary_exp_batch_predict.json"))
						})
					})
				})
			})
		})

		ginkgo.Describe("1.2. Accessing router logs via Turing API", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("GET /projects/:projectId/routers/:routerId/logs?component_type=:type", func() {
					for _, componentType := range []string{"router", "ensembler", "enricher"} {
						ginkgo.When(fmt.Sprintf("API is called to fetch %s's logs", componentType), func() {
							ginkgo.It(fmt.Sprintf("responds with %s's log entries", componentType), func() {
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

		ginkgo.Describe("1.3. Updating router with a new version, that can't be deployed", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("PUT /projects/:projectId/routers/:routerId", func() {
					ginkgo.When("request valid", func() {
						ginkgo.It("creates a new router's version", func() {
							apiE.PUT("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								WithJSON(JSONPayload("testdata/update_router_high_cpu.json.tmpl", cfg)).
								Expect().Status(http.StatusOK).
								JSON().Object()
						})
					})
				})

				ginkgo.Context("GET /projects/:projectId/routers/:routerId/versions/:version", func() {
					ginkgo.When("new router version can't be deployed", func() {
						ginkgo.It("fails after a timeout", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 2)

								g.Expect(version.Value("status").Raw()).ShouldNot(gomega.Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(gomega.Succeed())

							version.
								HasValue("status", api.Status.Failed).
								HasValue("error", "Requested CPU is more than max permissible")
						})

						ginkgo.It("keeps previously deployed version active", func() {
							router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)
							router.Path("$.config.version").IsEqual(1)
						})
					})
				})
			})

			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {
					ginkgo.When("valid request", func() {
						ginkgo.It("responds with expected payload", func() {
							gomega.Consistently(func(gomega.Gomega) {
								AssertPredictResponse()
							}, arbitraryUpdateIntervals...).Should(gomega.Succeed())
						})
					})
				})
			})
		})

		ginkgo.Describe("1.4. Undeploying router", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.When("router is deployed", func() {
					ginkgo.Context("POST /projects/:projectId/routers/:routerId/undeploy", func() {
						ginkgo.It("accepts request and starts undeploying the router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/undeploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)
						})

						ginkgo.It("eventually undeploys the router", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(gomega.And(
									gomega.Not(gomega.HaveKey("endpoint")),
									gomega.HaveKeyWithValue("status", api.Status.Undeployed),
									gomega.HaveKeyWithValue("config", gomega.And(
										gomega.HaveKeyWithValue("version", gomega.BeNumerically("==", 1)),
										gomega.HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, defaultDeletionIntervals...).Should(gomega.Succeed())
						})
					})
				})
			})
		})

		ginkgo.Describe("1.5. Deploying a version with invalid configuration", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("POST /projects/:projectId/routers/:routerId/versions/:version/deploy", func() {
					ginkgo.When("version exists", func() {
						ginkgo.It("attempts to deploy this version", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/versions/{version}/deploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								WithPath("version", version.Value("version").Raw()).
								Expect().Status(http.StatusAccepted)

							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)
								g.Expect(router.Value("status").Raw()).To(gomega.Equal(api.Status.Undeployed))
							}, arbitraryUpdateIntervals...).Should(gomega.Succeed())
						})

						ginkgo.It("fails after a timeout", func() {
							versionID := version.Value("version").Raw()
							gomega.Eventually(func(g gomega.Gomega) {
								version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, versionID)

								g.Expect(version.Value("status").Raw()).ShouldNot(gomega.Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(gomega.Succeed())

							version.
								HasValue("status", api.Status.Failed).
								HasValue("error", "Requested CPU is more than max permissible")
						})

						ginkgo.It("keeps previous valid version as router's current version", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(gomega.And(
									gomega.Not(gomega.HaveKey("endpoint")),
									gomega.HaveKeyWithValue("status", api.Status.Undeployed),
									gomega.HaveKeyWithValue("config", gomega.And(
										gomega.HaveKeyWithValue("version", gomega.BeNumerically("==", 1)),
										gomega.HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, arbitraryUpdateIntervals...).Should(gomega.Succeed())
						})
					})
				})
			})
		})

		ginkgo.Describe("1.6. Redeploying a router with existing valid configuration", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("POST /projects/:projectId/routers/:routerId/deploy", func() {
					ginkgo.When("router is not deployed", func() {
						ginkgo.It("attempts to deploy the router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/deploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusAccepted)

							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Value("status").Raw()).To(gomega.Equal(api.Status.Pending))
							}, arbitraryUpdateIntervals...).Should(gomega.Succeed())
						})

						ginkgo.It("successfully deploys the router", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Value("status").Raw()).ShouldNot(gomega.Equal(api.Status.Pending))
							}, defaultDeploymentIntervals...).Should(gomega.Succeed())

							router.
								HasValue("status", api.Status.Deployed).
								Path("$.config.version").IsEqual(1)
						})
					})
				})
			})
		})

		ginkgo.Describe("1.7. Updating router with a new valid configuration", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("PUT /projects/:projectId/routers/:routerId", func() {
					ginkgo.When("request valid", func() {
						ginkgo.It("creates a new router's version", func() {
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

				ginkgo.It("successfully deploys a new version", func() {
					gomega.Eventually(func(g gomega.Gomega) {
						version = api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 3)

						g.Expect(version.Value("status").Raw()).ShouldNot(gomega.Equal(api.Status.Pending))
					}, defaultDeploymentIntervals...).Should(gomega.Succeed())

					version.
						HasValue("status", api.Status.Deployed).
						NotContainsKey("error")
				})

				ginkgo.It("marks previous version as undeployed", func() {
					gomega.Eventually(func(g gomega.Gomega) {
						v := api.GetRouterVersion(apiE, routerCtx.ProjectID, routerCtx.ID, 1)

						g.Expect(v.Value("status").Raw()).Should(gomega.Equal(api.Status.Undeployed))
					}, arbitraryUpdateIntervals...).Should(gomega.Succeed())
				})

				ginkgo.It("updates router's configuration to the new version", func() {
					api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID).
						HasValue("status", api.Status.Deployed).
						Path("$.config.version").IsEqual(3)
				})
			})

			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {
					ginkgo.When("when the new version of router is deployed", func() {
						ginkgo.It("responds with expected payload", func() {
							want = httpexpect.NewValue(
								ginkgo.GinkgoT(),
								json.RawMessage(`{"version": "treatment-a"}`),
							).Object()
							AssertPredictResponse()
						})
					})
				})
			})
		})

		ginkgo.Describe("1.8. Deleting router", func() {
			ginkgo.Context("Turing API", func() {
				ginkgo.Context("DELETE /projects/:projectId/routers/:routerId", func() {
					ginkgo.When("router is deployed", func() {
						ginkgo.It("responds with bad request status", func() {
							apiE.DELETE("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusBadRequest).
								JSON().
								IsEqual(json.RawMessage(`{
                                        "description": "invalid delete request",
                                        "error": "router is currently deployed. Undeploy it first."
                                    }`))
						})
					})
				})

				ginkgo.Context("POST /projects/:projectId/routers/:routerId/undeploy", func() {
					ginkgo.When("router is deployed", func() {
						ginkgo.It("accepts request for undeploying router", func() {
							apiE.POST("/projects/{projectId}/routers/{routerId}/undeploy").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)
						})

						ginkgo.It("undeploys the router", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								router = api.GetRouter(apiE, routerCtx.ProjectID, routerCtx.ID)

								g.Expect(router.Raw()).To(gomega.And(
									gomega.Not(gomega.HaveKey("endpoint")),
									gomega.HaveKeyWithValue("status", api.Status.Undeployed),
									gomega.HaveKeyWithValue("config", gomega.And(
										gomega.HaveKeyWithValue("version", gomega.BeNumerically("==", 3)),
										gomega.HaveKeyWithValue("status", api.Status.Undeployed))),
								))
							}, defaultDeletionIntervals...).Should(gomega.Succeed())
						})
					})
				})

				ginkgo.Context("DELETE /projects/:projectId/routers/:routerId", func() {
					ginkgo.When("router is not deployed", func() {
						ginkgo.It("successfully deletes router from the db", func() {
							apiE.DELETE("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusOK)

							apiE.GET("/projects/{projectId}/routers/{routerId}").
								WithPath("projectId", routerCtx.ProjectID).
								WithPath("routerId", routerCtx.ID).
								Expect().Status(http.StatusNotFound).
								JSON().
								IsEqual(json.RawMessage(`{
									"description": "router not found",
									"error": "record not found"
								}`))
						})
					})
				})
			})

			ginkgo.Context("Turing Router API", func() {
				ginkgo.Context("POST /v1/predict", func() {
					ginkgo.When("when the router is deleted", func() {
						ginkgo.It("responds with NotFound status code", func() {
							gomega.Eventually(func(g gomega.Gomega) {
								response := routerE.POST("/v1/predict").
									WithHeaders(defaultPredictHeaders).
									WithJSON(json.RawMessage(`{}`)).
									Expect()

								g.Expect(response.Raw().StatusCode).To(gomega.Equal(http.StatusNotFound))
							}, defaultDeletionIntervals...).Should(gomega.Succeed())
						})
					})
				})
			})

			ginkgo.Context("K8s API", func() {
				ginkgo.When("the router is deleted", func() {
					ginkgo.It("should remove all its k8s resources from the cluster", func() {
						routerName := router.Value("name").String().Raw()
						gomega.Eventually(func(g gomega.Gomega) {
							k8sResources, err := cluster.ListRouterResources(cfg.Project.Name, routerName)

							g.Expect(err).ShouldNot(gomega.HaveOccurred())
							g.Expect(k8sResources).Should(gomega.And(
								gomega.HaveField("KnativeServices", gomega.BeEmpty()),
								gomega.HaveField("K8sServices", gomega.BeEmpty()),
								gomega.HaveField("IstioServices", gomega.BeEmpty()),
								gomega.HaveField("K8sDeployments", gomega.BeEmpty()),
								gomega.HaveField("ConfigMaps", gomega.BeEmpty()),
								gomega.HaveField("Secrets", gomega.BeEmpty()),
								gomega.HaveField("PVCs", gomega.BeEmpty()),
							))
						}, defaultDeletionIntervals...).Should(gomega.Succeed())
					})
				})
			})
		})
	})
