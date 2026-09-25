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

func TestPopulateXYMetricAxisDefault_injectsYWhenOmitted(t *testing.T) {
	t.Parallel()

	model := populateXYMetricAxisDefault(map[string]any{
		"operation":     "count",
		"empty_as_null": true,
	})

	assert.Equal(t, "y", model["axis"])
}

func TestPopulateXYMetricAxisDefault_preservesExplicitAxis(t *testing.T) {
	t.Parallel()

	model := populateXYMetricAxisDefault(map[string]any{
		"operation":     "count",
		"empty_as_null": true,
		"axis":          "y2",
	})

	assert.Equal(t, "y2", model["axis"])
}

func TestPopulateXYMetricDefaults_countInjectsEmptyAsNullAndAxis(t *testing.T) {
	t.Parallel()

	model := PopulateXYMetricDefaults(map[string]any{"operation": "count"})

	assert.Equal(t, false, model["empty_as_null"])
	assert.Equal(t, "y", model["axis"])
}

func TestPopulateXYMetricDefaults_preservesPractitionerColorWhenInjectingAxis(t *testing.T) {
	t.Parallel()

	static := map[string]any{"type": "static", "color": "#54B399"}
	model := PopulateXYMetricDefaults(map[string]any{
		"operation": "count",
		"color":     static,
	})

	assert.Equal(t, static, model["color"])
	assert.Equal(t, "y", model["axis"])
}

func TestPopulateXYMetricDefaults_percentileOmitsEmptyAsNullAndStillInjectsAxis(t *testing.T) {
	t.Parallel()

	model := PopulateXYMetricDefaults(map[string]any{
		"operation":  "percentile",
		"field":      "bytes",
		"percentile": float64(95),
	})

	_, hasEmptyAsNull := model["empty_as_null"]
	assert.False(t, hasEmptyAsNull)
	assert.Equal(t, "y", model["axis"])
}

func TestPopulateDatatableMetricDefaults_injectsVisibleAlignmentAndColorWithoutAxis(t *testing.T) {
	t.Parallel()

	model := PopulateDatatableMetricDefaults(map[string]any{"operation": "count"})

	assert.Equal(t, true, model["visible"])
	assert.Equal(t, "right", model["alignment"])
	assert.Equal(t, map[string]any{"type": "auto"}, model["color"])
	_, hasAxis := model["axis"]
	assert.False(t, hasAxis)
}

func TestPopulateLegacyMetricMetricDefaults_injectsColorAndSizeWithoutAxis(t *testing.T) {
	t.Parallel()

	model := PopulateLegacyMetricMetricDefaults(map[string]any{"operation": "count"})

	assert.Equal(t, map[string]any{"type": "auto"}, model["color"])
	assert.Equal(t, "m", model["size"])
	_, hasAxis := model["axis"]
	assert.False(t, hasAxis)
}

func TestPopulateLensMetricDefaults_doesNotInjectAxis(t *testing.T) {
	t.Parallel()

	model := PopulateLensMetricDefaults(map[string]any{"operation": "count"})

	_, hasAxis := model["axis"]
	assert.False(t, hasAxis)
}

func TestPopulateLegacyMetricMetricDefaults_preservesExplicitColorAndSize(t *testing.T) {
	t.Parallel()

	model := PopulateLegacyMetricMetricDefaults(map[string]any{
		"operation": "count",
		"color":     map[string]any{"type": "static", "color": "#54B399"},
		"size":      "l",
	})

	assert.Equal(t, map[string]any{"type": "static", "color": "#54B399"}, model["color"])
	assert.Equal(t, "l", model["size"])
}

func TestPopulateDatatableMetricDefaults_preservesExplicitVisibleAndAlignment(t *testing.T) {
	t.Parallel()

	model := PopulateDatatableMetricDefaults(map[string]any{
		"operation": "count",
		"visible":   false,
		"alignment": "left",
	})

	assert.Equal(t, false, model["visible"])
	assert.Equal(t, "left", model["alignment"])
}
