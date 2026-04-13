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

package dashboard

import (
	"math"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func apiTimeSliderConfig(start, end *float32, anchored *bool) struct {
	EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
	IsAnchored                 *bool    `json:"is_anchored,omitempty"`
	StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
} {
	return struct {
		EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
		IsAnchored                 *bool    `json:"is_anchored,omitempty"`
		StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
	}{
		StartPercentageOfTimeRange: start,
		EndPercentageOfTimeRange:   end,
		IsAnchored:                 anchored,
	}
}

// Test: when existing config block is nil and API returns empty config, leave nil.
func Test_populateTimeSliderControlFromAPI_nilBlock_emptyAPIConfig(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{}
	populateTimeSliderControlFromAPI(pm, tfPanel, apiTimeSliderConfig(nil, nil, nil))
	assert.Nil(t, pm.TimeSliderControlConfig)
}

// Test: when existing config block is nil and API returns data, preserve nil (null-preservation).
func Test_populateTimeSliderControlFromAPI_nilBlock_withAPIData(t *testing.T) {
	pm := &panelModel{}
	tfPanel := &panelModel{}
	populateTimeSliderControlFromAPI(pm, tfPanel, apiTimeSliderConfig(new(float32(0.1)), new(float32(0.9)), new(true)))
	assert.Nil(t, pm.TimeSliderControlConfig)
}

// Test: on import (tfPanel == nil) with API data, populate block from API.
func Test_populateTimeSliderControlFromAPI_import_withAPIData(t *testing.T) {
	pm := &panelModel{}
	populateTimeSliderControlFromAPI(pm, nil, apiTimeSliderConfig(new(float32(0.1)), new(float32(0.9)), new(true)))
	require.NotNil(t, pm.TimeSliderControlConfig)
	assert.Equal(t, types.Float32Value(0.1), pm.TimeSliderControlConfig.StartPercentageOfTimeRange)
	assert.Equal(t, types.Float32Value(0.9), pm.TimeSliderControlConfig.EndPercentageOfTimeRange)
	assert.Equal(t, types.BoolValue(true), pm.TimeSliderControlConfig.IsAnchored)
}

// Test: on import (tfPanel == nil) with empty API config, leave nil.
func Test_populateTimeSliderControlFromAPI_import_emptyAPIConfig(t *testing.T) {
	pm := &panelModel{}
	populateTimeSliderControlFromAPI(pm, nil, apiTimeSliderConfig(nil, nil, nil))
	assert.Nil(t, pm.TimeSliderControlConfig)
}

// Test: when config block exists with known fields, populate from API.
func Test_populateTimeSliderControlFromAPI_knownFields_populatedFromAPI(t *testing.T) {
	pm := &panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.1),
			EndPercentageOfTimeRange:   types.Float32Value(0.9),
			IsAnchored:                 types.BoolValue(false),
		},
	}
	tfPanel := &panelModel{TimeSliderControlConfig: pm.TimeSliderControlConfig}
	populateTimeSliderControlFromAPI(pm, tfPanel, apiTimeSliderConfig(new(float32(0.2)), new(float32(0.8)), new(true)))
	require.NotNil(t, pm.TimeSliderControlConfig)
	assert.Equal(t, types.Float32Value(0.2), pm.TimeSliderControlConfig.StartPercentageOfTimeRange)
	assert.Equal(t, types.Float32Value(0.8), pm.TimeSliderControlConfig.EndPercentageOfTimeRange)
	assert.Equal(t, types.BoolValue(true), pm.TimeSliderControlConfig.IsAnchored)
}

// Test: null-preservation — null fields in state are not overwritten by API values.
func Test_populateTimeSliderControlFromAPI_nullFields_preservedAsNull(t *testing.T) {
	pm := &panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Null(),
			EndPercentageOfTimeRange:   types.Float32Null(),
			IsAnchored:                 types.BoolNull(),
		},
	}
	tfPanel := &panelModel{TimeSliderControlConfig: pm.TimeSliderControlConfig}
	populateTimeSliderControlFromAPI(pm, tfPanel, apiTimeSliderConfig(new(float32(0.5)), new(float32(0.5)), new(true)))
	require.NotNil(t, pm.TimeSliderControlConfig)
	assert.True(t, pm.TimeSliderControlConfig.StartPercentageOfTimeRange.IsNull())
	assert.True(t, pm.TimeSliderControlConfig.EndPercentageOfTimeRange.IsNull())
	assert.True(t, pm.TimeSliderControlConfig.IsAnchored.IsNull())
}

// Test: mixed — some fields known, some null.
func Test_populateTimeSliderControlFromAPI_mixedFields(t *testing.T) {
	pm := &panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.1),
			EndPercentageOfTimeRange:   types.Float32Null(),
			IsAnchored:                 types.BoolNull(),
		},
	}
	tfPanel := &panelModel{TimeSliderControlConfig: pm.TimeSliderControlConfig}
	populateTimeSliderControlFromAPI(pm, tfPanel, apiTimeSliderConfig(new(float32(0.2)), new(float32(0.8)), new(true)))
	require.NotNil(t, pm.TimeSliderControlConfig)
	// known field is updated
	assert.Equal(t, types.Float32Value(0.2), pm.TimeSliderControlConfig.StartPercentageOfTimeRange)
	// null fields are preserved
	assert.True(t, pm.TimeSliderControlConfig.EndPercentageOfTimeRange.IsNull())
	assert.True(t, pm.TimeSliderControlConfig.IsAnchored.IsNull())
}

// Test: buildTimeSliderControlConfig sets known fields and omits null fields.
func Test_buildTimeSliderControlConfig_knownFields(t *testing.T) {
	pm := panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.25),
			EndPercentageOfTimeRange:   types.Float32Value(0.75),
			IsAnchored:                 types.BoolValue(true),
		},
	}
	tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
		Config: struct {
			EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
			IsAnchored                 *bool    `json:"is_anchored,omitempty"`
			StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
		}{},
	}
	buildTimeSliderControlConfig(pm, &tsPanel)
	require.NotNil(t, tsPanel.Config.StartPercentageOfTimeRange)
	require.NotNil(t, tsPanel.Config.EndPercentageOfTimeRange)
	require.NotNil(t, tsPanel.Config.IsAnchored)
	// Exact float32 equality via IEEE 754 bits (build path copies ValueFloat32() without widening).
	require.Equal(t, math.Float32bits(float32(0.25)), math.Float32bits(*tsPanel.Config.StartPercentageOfTimeRange))
	require.Equal(t, math.Float32bits(float32(0.75)), math.Float32bits(*tsPanel.Config.EndPercentageOfTimeRange))
	assert.True(t, *tsPanel.Config.IsAnchored)
}

// Test: buildTimeSliderControlConfig omits null fields.
func Test_buildTimeSliderControlConfig_nullFields(t *testing.T) {
	pm := panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Null(),
			EndPercentageOfTimeRange:   types.Float32Null(),
			IsAnchored:                 types.BoolNull(),
		},
	}
	tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
		Config: struct {
			EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
			IsAnchored                 *bool    `json:"is_anchored,omitempty"`
			StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
		}{},
	}
	buildTimeSliderControlConfig(pm, &tsPanel)
	assert.Nil(t, tsPanel.Config.StartPercentageOfTimeRange)
	assert.Nil(t, tsPanel.Config.EndPercentageOfTimeRange)
	assert.Nil(t, tsPanel.Config.IsAnchored)
}

// Test: boundary values 0.0 and 1.0 are valid.
func Test_buildTimeSliderControlConfig_boundaryValues(t *testing.T) {
	pm := panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.0),
			EndPercentageOfTimeRange:   types.Float32Value(1.0),
			IsAnchored:                 types.BoolNull(),
		},
	}
	tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
		Config: struct {
			EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
			IsAnchored                 *bool    `json:"is_anchored,omitempty"`
			StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
		}{},
	}
	buildTimeSliderControlConfig(pm, &tsPanel)
	require.NotNil(t, tsPanel.Config.StartPercentageOfTimeRange)
	require.NotNil(t, tsPanel.Config.EndPercentageOfTimeRange)
	require.Equal(t, math.Float32bits(float32(0.0)), math.Float32bits(*tsPanel.Config.StartPercentageOfTimeRange))
	require.Equal(t, math.Float32bits(float32(1.0)), math.Float32bits(*tsPanel.Config.EndPercentageOfTimeRange))
}

// Test: non-dyadic decimals round-trip through API struct without float64 widening drift.
func Test_timeSliderPercentage_float32RoundTrip_writeThenRead(t *testing.T) {
	start := float32(0.1)
	end := float32(0.9)
	pm := panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(start),
			EndPercentageOfTimeRange:   types.Float32Value(end),
			IsAnchored:                 types.BoolValue(false),
		},
	}
	tsPanel := kbapi.KbnDashboardPanelTypeTimeSliderControl{
		Config: struct {
			EndPercentageOfTimeRange   *float32 `json:"end_percentage_of_time_range,omitempty"`
			IsAnchored                 *bool    `json:"is_anchored,omitempty"`
			StartPercentageOfTimeRange *float32 `json:"start_percentage_of_time_range,omitempty"`
		}{},
	}
	buildTimeSliderControlConfig(pm, &tsPanel)
	require.NotNil(t, tsPanel.Config.StartPercentageOfTimeRange)
	require.NotNil(t, tsPanel.Config.EndPercentageOfTimeRange)

	out := &panelModel{
		TimeSliderControlConfig: &timeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(start),
			EndPercentageOfTimeRange:   types.Float32Value(end),
			IsAnchored:                 types.BoolValue(false),
		},
	}
	tfPanel := &panelModel{TimeSliderControlConfig: out.TimeSliderControlConfig}
	populateTimeSliderControlFromAPI(out, tfPanel, tsPanel.Config)

	require.NotNil(t, out.TimeSliderControlConfig)
	assert.Equal(t, types.Float32Value(start), out.TimeSliderControlConfig.StartPercentageOfTimeRange)
	assert.Equal(t, types.Float32Value(end), out.TimeSliderControlConfig.EndPercentageOfTimeRange)
}
