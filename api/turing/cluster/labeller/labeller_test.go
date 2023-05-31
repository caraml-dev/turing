package labeller

import "testing"

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
				"component":    {},
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
				"gojek.com/component":    {},
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
