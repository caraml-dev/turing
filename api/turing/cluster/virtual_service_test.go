package cluster

import (
	"testing"

	"gotest.tools/assert"
	networking "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
)

func TestBuildVirtualService(t *testing.T) {
	cfg := &VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: "test-namespace",
		Labels: map[string]string{
			"key": "value",
		},
		Gateway:          "gateway",
		Endpoint:         "test-svc-turing-router.models.example.com",
		DestinationHost:  "istio",
		HostRewrite:      "test-svc-turing-router-1.models.example.com",
		MatchURIPrefixes: []string{"/v1/prefix"},
	}
	expected := v1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: networking.VirtualService{
			Hosts:    []string{cfg.Endpoint},
			Gateways: []string{"gateway"},
			Http: []*networking.HTTPRoute{
				{
					Match: []*networking.HTTPMatchRequest{
						{
							Uri: &networking.StringMatch{
								MatchType: &networking.StringMatch_Prefix{
									Prefix: "/v1/prefix",
								},
							},
						},
					},
					Route: []*networking.HTTPRouteDestination{
						{
							Destination: &networking.Destination{
								Host: "istio",
							},
							Headers: &networking.Headers{
								Request: &networking.Headers_HeaderOperations{
									Set: map[string]string{"Host": cfg.HostRewrite},
								},
							},
							Weight: 100,
						},
					},
				},
			},
		},
	}
	got := cfg.BuildVirtualService()
	err := tu.CompareObjects(expected, *got)
	assert.NilError(t, err)
}
