package labeller

import (
	"testing"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/google/go-cmp/cmp"
)

func TestLabeller(t *testing.T) {
	tests := map[string]struct {
		doInit       bool
		prefix       string
		expectedKeys map[string]struct{}
	}{
		"success | no init called": {
			doInit: false,
			expectedKeys: map[string]struct{}{
				"environment":  {},
				"stream":       {},
				"team":         {},
				"app":          {},
				"orchestrator": {},
			},
		},
		"success | init called": {
			doInit: true,
			prefix: "gojek.com/",
			expectedKeys: map[string]struct{}{
				"gojek.com/environment":  {},
				"gojek.com/stream":       {},
				"gojek.com/team":         {},
				"gojek.com/app":          {},
				"gojek.com/orchestrator": {},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Always reset singleton objects.
			defer InitKubernetesLabeller("", "dev")

			if tt.doInit {
				InitKubernetesLabeller(tt.prefix, "dev")
			}

			labels := BuildLabels(KubernetesLabelsRequest{})
			for key := range tt.expectedKeys {
				if _, ok := labels[key]; !ok {
					t.Errorf("expected key %s", key)
				}
			}
		})
	}
}

func TestBuildLabels(t *testing.T) {
	tests := []struct {
		name     string
		request  KubernetesLabelsRequest
		expected map[string]string
	}{
		{
			name: "Test with valid request",
			request: KubernetesLabelsRequest{
				Stream: "testStream",
				Team:   "testTeam",
				App:    "testApp",
				Labels: []mlp.Label{
					{Key: "customLabel1", Value: "value1"},
					{Key: "customLabel2", Value: "value2"},
				},
			},
			expected: map[string]string{
				"orchestrator": "turing",
				"stream":       "testStream",
				"team":         "testTeam",
				"app":          "testApp",
				"environment":  "dev",
				"customLabel1": "value1",
				"customLabel2": "value2",
			},
		},
		{
			name: "Test with reserved key",
			request: KubernetesLabelsRequest{
				Stream: "testStream",
				Team:   "testTeam",
				App:    "testApp",
				Labels: []mlp.Label{
					{Key: "orchestrator", Value: "value1"}, // Reserved key
				},
			},
			expected: map[string]string{
				"orchestrator": "turing", // Should not be overridden
				"stream":       "testStream",
				"team":         "testTeam",
				"app":          "testApp",
				"environment":  "dev",
			},
		},
		{
			name: "Test with invalid label names",
			request: KubernetesLabelsRequest{
				Stream: "testStream",
				Team:   "testTeam",
				App:    "testApp",
				Labels: []mlp.Label{
					{Key: "invalid.Label", Value: "value1"}, // Invalid label key
				},
			},
			expected: map[string]string{
				"orchestrator": "turing",
				"stream":       "testStream",
				"team":         "testTeam",
				"app":          "testApp",
				"environment":  "dev",
			},
		},
	}

	for _, tt := range tests {
		// Always reset singleton objects.
		defer InitKubernetesLabeller("", "dev")
		t.Run(tt.name, func(t *testing.T) {
			got := BuildLabels(tt.request)
			if diff := cmp.Diff(got, tt.expected); diff != "" {
				t.Errorf("BuildLabels() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
