// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package lenscommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricFormatUsesUnsupportedDurationParams_decimalsWithoutCompact(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type":     "duration",
			"from":     "s",
			"to":       "ms",
			"decimals": float64(0),
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_decimalsUnderParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type": "duration",
			"params": map[string]any{
				"decimals": float64(0),
			},
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_compactUnderParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type": "duration",
			"params": map[string]any{
				"compact": true,
			},
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDSpelling(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id":       "duration",
			"decimals": float64(1),
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDCompactOnly(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id":      "duration",
			"compact": true,
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDDecimalsAndCompact(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id":       "duration",
			"decimals": float64(0),
			"compact":  true,
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDDecimalsUnderParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id": "duration",
			"params": map[string]any{
				"decimals": float64(1),
			},
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDCompactUnderParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id": "duration",
			"params": map[string]any{
				"compact": true,
			},
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_legacyIDWithoutParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"id":   "duration",
			"from": "s",
			"to":   "ms",
		},
	}

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_durationWithoutParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type": "duration",
			"from": "s",
			"to":   "ms",
		},
	}

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_numberFormat(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type":     "number",
			"decimals": float64(0),
			"compact":  true,
		},
	}

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_percentParams(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type": "percent",
			"params": map[string]any{
				"compact":  true,
				"decimals": float64(2),
			},
		},
	}

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_nullFormat(t *testing.T) {
	t.Parallel()

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(map[string]any{"format": nil}))
}

func TestMetricFormatUsesUnsupportedDurationParams_nilAndMissingFormat(t *testing.T) {
	t.Parallel()

	assert.False(t, MetricFormatUsesUnsupportedDurationParams(nil))
	assert.False(t, MetricFormatUsesUnsupportedDurationParams(map[string]any{}))
	assert.False(t, MetricFormatUsesUnsupportedDurationParams(map[string]any{"format": "duration"}))
}

func TestMetricFormatUsesUnsupportedDurationParams_decimalsAndCompact(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type":     "duration",
			"from":     "s",
			"to":       "ms",
			"decimals": float64(0),
			"compact":  true,
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}

func TestMetricFormatUsesUnsupportedDurationParams_compactWithoutDecimals(t *testing.T) {
	t.Parallel()

	model := map[string]any{
		"format": map[string]any{
			"type":    "duration",
			"from":    "s",
			"to":      "ms",
			"compact": true,
		},
	}

	assert.True(t, MetricFormatUsesUnsupportedDurationParams(model))
}
