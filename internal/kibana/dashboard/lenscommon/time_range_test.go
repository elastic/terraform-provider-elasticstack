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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeRangeStubResolver struct{}

func (timeRangeStubResolver) ResolveChartTimeRange(chartLevel *models.TimeRangeModel) *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	return ResolveChartTimeRange(nil, chartLevel)
}

func TestResolveChartTimeRange_nilWhenChartUnset(t *testing.T) {
	t.Parallel()
	assert.Nil(t, ResolveChartTimeRange(nil, nil))
}

func TestResolveChartTimeRange_returnsConfiguredChartLevel(t *testing.T) {
	t.Parallel()
	chart := &models.TimeRangeModel{
		From: types.StringValue("now-30d"),
		To:   types.StringValue("now-1d"),
		Mode: types.StringValue("relative"),
	}
	got := ResolveChartTimeRange(nil, chart)
	require.NotNil(t, got)
	assert.Equal(t, "now-30d", got.From)
	assert.Equal(t, "now-1d", got.To)
	require.NotNil(t, got.Mode)
	assert.Equal(t, "relative", string(*got.Mode))
}

func TestLensChartPresentationWritesFor_omitsTimeRangeWhenUnset(t *testing.T) {
	t.Parallel()
	writes, diags := LensChartPresentationWritesFor(timeRangeStubResolver{}, models.LensChartPresentationTFModel{})
	require.False(t, diags.HasError())
	assert.Nil(t, writes.TimeRange)
}

func TestLensChartPresentationWritesFor_setsTimeRangeWhenConfigured(t *testing.T) {
	t.Parallel()
	in := models.LensChartPresentationTFModel{
		TimeRange: &models.TimeRangeModel{
			From: types.StringValue("now-7d"),
			To:   types.StringValue("now"),
		},
	}
	writes, diags := LensChartPresentationWritesFor(timeRangeStubResolver{}, in)
	require.False(t, diags.HasError())
	require.NotNil(t, writes.TimeRange)
	assert.Equal(t, "now-7d", writes.TimeRange.From)
	assert.Equal(t, "now", writes.TimeRange.To)
}
