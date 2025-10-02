package datafeed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_populateScriptFieldsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "empty script fields model returns empty result",
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "script field with all defaults already set returns unchanged",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
						"params": map[string]any{
							"multiplier": 2,
						},
					},
					"ignore_failure": false,
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
						"params": map[string]any{
							"multiplier": 2,
						},
					},
					"ignore_failure": false,
				},
			},
		},
		{
			name: "script field missing ignore_failure gets default false",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
					},
					// ignore_failure is missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
					},
					"ignore_failure": false,
				},
			},
		},
		{
			name: "script field missing lang gets default painless",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						// lang is missing
					},
					"ignore_failure": true,
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
					},
					"ignore_failure": true,
				},
			},
		},
		{
			name: "script field with both missing defaults gets both set",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						// lang is missing
					},
					// ignore_failure is missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value * 2",
						"lang":   "painless",
					},
					"ignore_failure": false,
				},
			},
		},
		{
			name: "script field without script only gets ignore_failure default",
			input: map[string]any{
				"field1": map[string]any{
					"some_other_field": "value",
					// no script field, ignore_failure is missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"some_other_field": "value",
					"ignore_failure":   false,
				},
			},
		},
		{
			name: "multiple script fields get defaults independently",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field1'].value",
						// lang missing
					},
					"ignore_failure": true,
				},
				"field2": map[string]any{
					"script": map[string]any{
						"source": "doc['field2'].value",
						"lang":   "groovy",
					},
					// ignore_failure missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source": "doc['field1'].value",
						"lang":   "painless",
					},
					"ignore_failure": true,
				},
				"field2": map[string]any{
					"script": map[string]any{
						"source": "doc['field2'].value",
						"lang":   "groovy",
					},
					"ignore_failure": false,
				},
			},
		},
		{
			name: "preserves additional unknown fields",
			input: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source":        "doc['field'].value",
						"custom_option": "value",
						// lang missing
					},
					"custom_field": "custom_value",
					// ignore_failure missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script": map[string]any{
						"source":        "doc['field'].value",
						"custom_option": "value",
						"lang":          "painless",
					},
					"custom_field":   "custom_value",
					"ignore_failure": false,
				},
			},
		},
		{
			name: "handles non-map field values gracefully",
			input: map[string]any{
				"field1": "not a map",
				"field2": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value",
					},
				},
			},
			expected: map[string]any{
				"field1": "not a map",
				"field2": map[string]any{
					"script": map[string]any{
						"source": "doc['field'].value",
						"lang":   "painless",
					},
					"ignore_failure": false,
				},
			},
		},
		{
			name: "handles non-map script values gracefully",
			input: map[string]any{
				"field1": map[string]any{
					"script": "not a map",
					// ignore_failure missing
				},
			},
			expected: map[string]any{
				"field1": map[string]any{
					"script":         "not a map",
					"ignore_failure": false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := populateScriptFieldsDefaults(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
