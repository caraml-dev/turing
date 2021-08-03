package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genString(length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteString("a")
	}

	return sb.String()
}

func TestCheckDNSCompliantName(t *testing.T) {
	tests := map[string]struct {
		name     string
		padding  int
		expected bool
	}{
		"success | nominal": {
			name:     "a",
			padding:  0,
			expected: true,
		},
		"success | exactly maximum without padding": {
			name:     genString(63),
			padding:  0,
			expected: true,
		},
		"failure | exactly maximum +1 without padding": {
			name:     genString(64),
			padding:  0,
			expected: false,
		},
		"success | exactly maximum with padding": {
			name:     genString(43),
			padding:  20,
			expected: true,
		},
		"failure | exactly maximum +1 with padding": {
			name:     genString(44),
			padding:  20,
			expected: false,
		},
		"failure | empty": {
			name:     "",
			padding:  0,
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsQualifiedKubernetesName(tt.name, tt.padding)
			assert.Equal(t, tt.expected, result)
		})
	}
}
