package cluster

import (
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualService struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	Labels          map[string]string `json:"labels"`
	Gateway         string            `json:"gateway"`
	Endpoint        string            `json:"endpoint"`
	DestinationHost string            `json:"destination_host"`
	HostRewrite     string            `json:"host_rewrite"`
	MatchURIPrefix  []string          `json:"match_uri_prefix"`
}

func (cfg VirtualService) BuildVirtualService() *v1alpha3.VirtualService {
	httpRouteDest := &networking.HTTPRouteDestination{
		Destination: &networking.Destination{
			Host: cfg.DestinationHost,
		},
		Headers: &networking.Headers{
			Request: &networking.Headers_HeaderOperations{
				Set: map[string]string{"Host": cfg.HostRewrite},
			},
		},
		Weight: 100,
	}
	httpMatches := make([]*networking.HTTPMatchRequest, len(cfg.MatchURIPrefix))
	for index, prefix := range cfg.MatchURIPrefix {
		uri := networking.HTTPMatchRequest{
			Uri: &networking.StringMatch{
				MatchType: &networking.StringMatch_Prefix{
					Prefix: prefix,
				},
			},
		}
		httpMatches[index] = &uri
	}

	return &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: networking.VirtualService{
			Hosts:    []string{cfg.Endpoint},
			Gateways: []string{cfg.Gateway},
			Http: []*networking.HTTPRoute{
				{
					Match: httpMatches,
					Route: []*networking.HTTPRouteDestination{
						httpRouteDest,
					},
				},
			},
		},
	}
}
