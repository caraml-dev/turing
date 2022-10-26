package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"text/template"

	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gavv/httpexpect/v2"
	"github.com/mitchellh/mapstructure"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v2"

	"github.com/caraml-dev/turing/api/e2e/test/api"
	"github.com/caraml-dev/turing/api/e2e/test/cluster"
	"github.com/caraml-dev/turing/api/e2e/test/config"
)

var configFile string
var cfg config.Config

var defaultDeploymentIntervals = []interface{}{"10m", "5s"}
var defaultDeletionIntervals = []interface{}{"20s", "2s"}
var arbitraryUpdateIntervals = []interface{}{"10s", "1s"}

type TestData struct {
	config.Config
	TestContext interface{}
}

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "Path to a configuration file")
}

var istioVirtualServiceIntervals = []interface{}{"30s", "1s"}

var defaultPredictHeaders = map[string]string{
	"X-Mirror-Body": "true",
}

var _ = SynchronizedBeforeSuite(func() []byte {
	cfg, err := config.LoadFromFiles(configFile)
	Expect(err).NotTo(HaveOccurred())

	projects := config.NewHTTPExpect(GinkgoT(), cfg.APIBasePath).
		GET("/projects").
		Expect().
		Status(http.StatusOK).
		JSON().Array().NotEmpty()

	Expect(func() error {
		for _, p := range projects.Iter() {
			if p.Object().Path("$.name").String().Raw() == cfg.Project.Name {
				cfg.Project.ID = int(p.Object().Path("$.id").Number().Raw())
				return nil
			}
		}
		return fmt.Errorf(`project "%s" doesn't exist in Turing`, cfg.Project.Name)
	}()).To(Succeed())

	ensemblers := config.NewHTTPExpect(GinkgoT(), cfg.APIBasePath).
		GET("/projects/{projectId}/ensemblers").
		WithPath("projectId", cfg.Project.ID).
		Expect().Status(http.StatusOK).
		JSON().Object().
		Path("$.results").Array()

	Expect(func() error {
		wanted, got := []string{}, []string{}
		for _, pythonVersion := range cfg.PythonVersions {
			ensemblerName := fmt.Sprintf(`%s%s`, cfg.Ensemblers.BaseName, pythonVersion)
			wanted = append(wanted, ensemblerName)
			for _, e := range ensemblers.Iter() {
				if e.Object().Path("$.name").String().Raw() == ensemblerName {
					cfg.Ensemblers.Entities = append(cfg.Ensemblers.Entities,
						config.EnsemblerData{
							PythonVersion: pythonVersion,
							EnsemblerID:   int(e.Object().Path("$.id").Number().Raw())},
					)
					got = append(got, ensemblerName)
					break
				}
			}
		}
		if len(wanted) != len(got) {
			return fmt.Errorf(`Not all ensemblers were found. Wanted: %v, got: %v`, wanted, got)
		}
		return nil
	}()).To(Succeed())
	Expect(cluster.InitClusterClients(cfg)).To(Succeed())

	cfgYAML, err := yaml.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())

	return cfgYAML
}, func(cfgYAML []byte) {
	Expect(yaml.Unmarshal(cfgYAML, &cfg)).To(Succeed())
	Expect(cluster.InitClusterClients(&cfg)).To(Succeed())
})

func TestEndToEnd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Router Suite")
}

func DeployedRouterContext(payloadTpl string, protocol routerConfig.Protocol, args ...interface{}) bool {
	if len(args) == 0 {
		return false
	}

	var (
		body          func(ctx *RouterContext)
		ok            bool
		pythonVersion string
		testCaseName  string
	)
	if body, ok = args[len(args)-1].(func(ctx *RouterContext)); !ok {
		panic("last argument must have type func(ctx *RouterContext)")
	}
	if pythonVersion, ok = args[0].(string); ok {
		args = args[1 : len(args)-1]
		testCaseName = fmt.Sprintf("python-%s:%s", pythonVersion, payloadTpl)
	} else {
		args = args[:len(args)-1]
		testCaseName = payloadTpl
	}

	return Context("Turing", append(args, Ordered, func() {
		var (
			e         *httpexpect.Expect
			routerCtx RouterContext
			router    *httpexpect.Object
		)

		BeforeAll(func() {
			e = config.NewHTTPExpect(GinkgoT(), cfg.APIBasePath)

			var ensembler config.EnsemblerData
			if pythonVersion != "" {
				for _, e := range cfg.Ensemblers.Entities {
					if strings.HasPrefix(e.PythonVersion, pythonVersion) {
						ensembler = config.EnsemblerData{
							PythonVersion: strings.ReplaceAll(e.PythonVersion, ".", "-"),
							EnsemblerID:   e.EnsemblerID,
						}
						break
					}
				}
				Expect(ensembler.EnsemblerID).ShouldNot(Equal(0))
			}

			testData := TestData{cfg, ensembler}
			router = e.POST("/projects/{projectID}/routers", cfg.Project.ID).
				WithJSON(JSONPayload(payloadTpl, testData)).
				Expect().Status(http.StatusOK).
				JSON().Object()

			DeferCleanup(func() {
				Expect(
					cluster.CleanupRouterDeployment(cfg.Project.Name, router.Value("name").String().Raw()),
				).To(Succeed())
			})

			routerID := int(router.Value("id").Number().Raw())

			Eventually(func(g Gomega) {
				router = api.GetRouter(e, cfg.Project.ID, routerID)

				g.Expect(router.Value("status").Raw()).ShouldNot(Equal(api.Status.Pending))
			}, defaultDeploymentIntervals...).Should(Succeed())

			router.
				ValueEqual("status", "deployed").
				Value("config").Object().ValueEqual("version", 1)

			endpoint, err := url.Parse(router.Value("endpoint").String().Raw())
			Expect(err).ShouldNot(HaveOccurred())

			endpoint.Path = "/"

			if protocol == routerConfig.UPI {
				routerCtx = RouterContext{
					ID:        routerID,
					ProjectID: cfg.Project.ID,
					Endpoint:  router.Value("endpoint").String().Raw(),
				}
			} else {
				endpoint, err := url.Parse(router.Value("endpoint").String().Raw())
				Expect(err).ShouldNot(HaveOccurred())

				endpoint.Path = "/"

				routerCtx = RouterContext{
					ID:        routerID,
					ProjectID: cfg.Project.ID,
					Endpoint:  endpoint.String(),
				}
			}

		})

		When(fmt.Sprintf("%s router is deployed", testCaseName), func() {
			// Istio VirtualService configuration is applied asynchronously, so the fact that it
			// exists in the cluster doesn't mean that it was already being in use.
			// In the newer version of Istio (starting from v1.6), it is possible to wait for the
			// changes readiness: https://istio.io/latest/docs/ops/configuration/mesh/config-resource-ready/
			// However, we are currently on Istio v1.3 and there is no straight-forward
			// way of achieving this readiness check, hence we just add an arbitrary sleep to give
			// Istio some time to apply the changes.
			//
			// TODO: Remove once Turing is migrated on a newer version of Istio
			When("virtual service configuration is applied", func() {
				It("responds with a status, that is not 404 NotFound", func() {
					Eventually(func(g Gomega) {
						if protocol == routerConfig.UPI {
							conn, _ := grpc.Dial(routerCtx.Endpoint,
								grpc.WithTransportCredentials(insecure.NewCredentials()))
							defer conn.Close()

							client := upiv1.NewUniversalPredictionServiceClient(conn)
							upiRequest := &upiv1.PredictValuesRequest{}
							headers := metadata.New(map[string]string{"region": "region-a"})
							_, err := client.PredictValues(metadata.NewOutgoingContext(context.Background(), headers),
								upiRequest)

							g.Expect(err).To(BeNil())
						} else {
							resp := config.NewHTTPExpect(GinkgoT(), routerCtx.Endpoint).
								GET("/v1/predict").
								Expect().Raw()

							if resp.Body != nil {
								defer resp.Body.Close()
							}

							g.Expect(resp.StatusCode).NotTo(Equal(http.StatusNotFound))
						}
					}, istioVirtualServiceIntervals...).Should(Succeed())
				})
			})

			body(&routerCtx)
		})
	})...)
}

type RouterContext struct {
	ID        int
	ProjectID int
	Endpoint  string
}

func JSONPayload(tplPath string, args ...interface{}) json.RawMessage {
	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		panic(err)
	}

	var data interface{}
	if len(args) > 0 {
		data = args[0]
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func AssertResponsePayload(want, got *httpexpect.Object) {
	type ensemblerResponse struct {
		Response struct {
			RouteResponses []interface{} `mapstructure:"route_responses"`
		} `mapstructure:"response"`
	}

	var resp ensemblerResponse
	_ = mapstructure.Decode(want.Raw(), &resp)

	if len(resp.Response.RouteResponses) > 0 {
		got.
			ValueEqual("request", want.Value("request").Raw()).
			Value("response").Object().
			ValueEqual("experiment", want.Path("$.response.experiment").Raw()).
			Value("route_responses").Array().
			ContainsOnly(want.Path("$.response.route_responses").Array().Raw()...)
	} else {
		got.Equal(want.Raw())
	}
}
