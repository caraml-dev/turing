package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	originalMap := map[string]interface{}{
		"list_to_override": []interface{}{
			"item1",
		},
		"hello": map[string]interface{}{
			"world": "hello",
			"deep": map[string]interface{}{
				"purple": "band",
			},
			"integer": 5,
			"list": []interface{}{
				"string",
				1,
			},
		},
	}
	var tests = map[string]struct {
		override map[string]interface{}
		expected map[string]interface{}
	}{
		"nominal | no override": {
			override: map[string]interface{}{},
			expected: originalMap,
		},
		"nominal | override single nested key": {
			override: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": "override",
				},
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"item1",
				},
				"hello": map[string]interface{}{
					"world": "override",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"string",
						1,
					},
				},
			},
		},
		"nominal | add key at base level": {
			override: map[string]interface{}{
				"newKey": "newValue",
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"item1",
				},
				"newKey": "newValue",
				"hello": map[string]interface{}{
					"world": "hello",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"string",
						1,
					},
				},
			},
		},
		"nominal | add key at one level deep": {
			override: map[string]interface{}{
				"hello": map[string]interface{}{
					"newKey": "newValue",
				},
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"item1",
				},
				"hello": map[string]interface{}{
					"newKey": "newValue",
					"world":  "hello",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"string",
						1,
					},
				},
			},
		},
		"nominal | replace list": {
			override: map[string]interface{}{
				"hello": map[string]interface{}{
					"list": []interface{}{
						"override",
						2,
						"manual override",
					},
				},
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"item1",
				},
				"hello": map[string]interface{}{
					"world": "hello",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"override",
						2,
						"manual override",
					},
				},
			},
		},
		"nominal | add list base level": {
			override: map[string]interface{}{
				"list": []interface{}{
					"override",
					2,
					"manual override",
				},
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"item1",
				},
				"list": []interface{}{
					"override",
					2,
					"manual override",
				},
				"hello": map[string]interface{}{
					"world": "hello",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"string",
						1,
					},
				},
			},
		},
		"nominal | replace_list_base_level": {
			override: map[string]interface{}{
				"list_to_override": []interface{}{
					"override",
					2,
					"manual override",
				},
			},
			expected: map[string]interface{}{
				"list_to_override": []interface{}{
					"override",
					2,
					"manual override",
				},
				"hello": map[string]interface{}{
					"world": "hello",
					"deep": map[string]interface{}{
						"purple": "band",
					},
					"integer": 5,
					"list": []interface{}{
						"string",
						1,
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			merged, err := MergeMaps(originalMap, tt.override)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, merged)
		})
	}
}
