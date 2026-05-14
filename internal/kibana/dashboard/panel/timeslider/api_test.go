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

package timeslider_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/timeslider"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestContract_TimeSliderPanel(t *testing.T) {
	t.Parallel()

	contracttest.Run(t, timeslider.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "time_slider_control",
			"grid": {"x": 1, "y": 2, "w": 12, "h": 6},
			"id": "ts-contract",
			"config": {
				"start_percentage_of_time_range": 0.1,
				"end_percentage_of_time_range": 0.95,
				"is_anchored": false
			}
		}`,
	})
}

func TestPinned_TimeSliderPinnedRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ph := timeslider.Handler{}.PinnedHandler()

	in := models.PinnedPanelModel{
		Type: types.StringValue("time_slider_control"),
		TimeSliderControlConfig: &models.TimeSliderControlConfigModel{
			StartPercentageOfTimeRange: types.Float32Value(0.2),
			EndPercentageOfTimeRange:   types.Float32Value(0.85),
			IsAnchored:                 types.BoolValue(true),
		},
	}

	raw, d1 := ph.ToAPI(in)
	require.False(t, d1.HasError(), "%v", d1)

	out, d2 := ph.FromAPI(ctx, nil, raw)
	require.False(t, d2.HasError(), "%v", d2)

	require.True(t, out.Type.Equal(types.StringValue("time_slider_control")))
	require.NotNil(t, out.TimeSliderControlConfig)
	got := out.TimeSliderControlConfig
	require.True(t, got.StartPercentageOfTimeRange.Equal(in.TimeSliderControlConfig.StartPercentageOfTimeRange))
	require.True(t, got.EndPercentageOfTimeRange.Equal(in.TimeSliderControlConfig.EndPercentageOfTimeRange))
	require.True(t, got.IsAnchored.Equal(in.TimeSliderControlConfig.IsAnchored))
	discriminator, err := raw.Discriminator()
	require.NoError(t, err)
	require.Equal(t, "time_slider_control", discriminator)
}
