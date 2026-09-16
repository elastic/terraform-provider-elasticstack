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

package lensxy

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func axisCfg[D any, S ~string](domain *D, scale S, grid, ticks bool, orient kbapi.KibanaHTTPAPIsVisApiOrientation, title string, titleVis bool) *axisConfigAPIModel[D, S] {
	return &axisConfigAPIModel[D, S]{
		Domain: domain,
		Grid: &struct {
			Visible bool `json:"visible"`
		}{Visible: grid},
		Labels: &struct {
			Orientation *kbapi.KibanaHTTPAPIsVisApiOrientation `json:"orientation,omitempty"`
		}{Orientation: &orient},
		Scale: &scale,
		Ticks: &struct {
			Visible bool `json:"visible"`
		}{Visible: ticks},
		Title: &struct {
			Text    *string `json:"text,omitempty"`
			Visible *bool   `json:"visible,omitempty"`
		}{Text: &title, Visible: &titleVis},
	}
}

func TestXyAxisFromAPIToAPI_XYY2RoundTrip(t *testing.T) {
	orientationAngled := kbapi.KibanaHTTPAPIsVisApiOrientationAngled
	orientationHorizontal := kbapi.KibanaHTTPAPIsVisApiOrientationHorizontal
	orientationVertical := kbapi.KibanaHTTPAPIsVisApiOrientationVertical
	xScale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScaleTemporal
	yScale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScaleSqrt
	y2Scale := kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2ScaleLog

	var xDomain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Domain
	require.NoError(t, xDomain.FromKibanaHTTPAPIsVisApiDomainCustom(kbapi.KibanaHTTPAPIsVisApiDomainCustom{
		Type: kbapi.KibanaHTTPAPIsVisApiDomainCustomTypeCustom, Min: 1, Max: 2,
	}))
	var yDomain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain
	require.NoError(t, yDomain.FromKibanaHTTPAPIsVisApiDomainCustom(kbapi.KibanaHTTPAPIsVisApiDomainCustom{
		Type: kbapi.KibanaHTTPAPIsVisApiDomainCustomTypeCustom, Min: 3, Max: 4,
	}))
	var y2Domain kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain
	require.NoError(t, y2Domain.FromKibanaHTTPAPIsVisApiDomainCustom(kbapi.KibanaHTTPAPIsVisApiDomainCustom{
		Type: kbapi.KibanaHTTPAPIsVisApiDomainCustomTypeCustom, Min: 5, Max: 6,
	}))

	apiAxis := &kbapi.KibanaHTTPAPIsVisApiXyAxisConfig{
		X:  axisCfg(&xDomain, xScale, true, true, orientationAngled, "X title", true),
		Y:  axisCfg(&yDomain, yScale, false, false, orientationHorizontal, "Y title", true),
		Y2: axisCfg(&y2Domain, y2Scale, true, false, orientationVertical, "Y2 title", false),
	}

	m := &models.XYAxisModel{}
	diags := xyAxisFromAPI(m, apiAxis)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, m.X)
	require.True(t, m.X.Grid.ValueBool())
	require.True(t, m.X.Ticks.ValueBool())
	require.Equal(t, "angled", m.X.LabelOrientation.ValueString())
	require.Equal(t, "temporal", m.X.Scale.ValueString())
	require.Equal(t, "X title", m.X.Title.Value.ValueString())
	require.Contains(t, m.X.DomainJSON.ValueString(), `"min":1`)
	require.Contains(t, m.X.DomainJSON.ValueString(), `"max":2`)

	require.NotNil(t, m.Y)
	require.False(t, m.Y.Grid.ValueBool())
	require.False(t, m.Y.Ticks.ValueBool())
	require.Equal(t, "horizontal", m.Y.LabelOrientation.ValueString())
	require.Equal(t, "sqrt", m.Y.Scale.ValueString())
	require.Equal(t, "Y title", m.Y.Title.Value.ValueString())
	require.Contains(t, m.Y.DomainJSON.ValueString(), `"min":3`)
	require.Contains(t, m.Y.DomainJSON.ValueString(), `"max":4`)

	require.NotNil(t, m.Y2)
	require.True(t, m.Y2.Grid.ValueBool())
	require.False(t, m.Y2.Ticks.ValueBool())
	require.Equal(t, "vertical", m.Y2.LabelOrientation.ValueString())
	require.Equal(t, "log", m.Y2.Scale.ValueString())
	require.Equal(t, "Y2 title", m.Y2.Title.Value.ValueString())
	require.Contains(t, m.Y2.DomainJSON.ValueString(), `"min":5`)
	require.Contains(t, m.Y2.DomainJSON.ValueString(), `"max":6`)

	out, diags := xyAxisToAPI(m)
	require.False(t, diags.HasError(), "%v", diags)

	require.NotNil(t, out.X)
	require.NotNil(t, out.X.Scale)
	require.Equal(t, xScale, *out.X.Scale)
	require.NotNil(t, out.X.Grid)
	require.True(t, out.X.Grid.Visible)
	require.NotNil(t, out.X.Labels)
	require.Equal(t, orientationAngled, *out.X.Labels.Orientation)

	require.NotNil(t, out.Y)
	require.NotNil(t, out.Y.Scale)
	require.Equal(t, yScale, *out.Y.Scale)

	require.NotNil(t, out.Y2)
	require.NotNil(t, out.Y2.Scale)
	require.Equal(t, y2Scale, *out.Y2.Scale)
}

func TestXyAxisFromAPI_EmptyConfigsAreNil(t *testing.T) {
	apiAxis := &kbapi.KibanaHTTPAPIsVisApiXyAxisConfig{
		X:  &axisConfigAPIModel[kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_X_Domain, kbapi.KibanaHTTPAPIsVisApiXyAxisConfigXScale]{},
		Y:  &axisConfigAPIModel[kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y_Domain, kbapi.KibanaHTTPAPIsVisApiXyAxisConfigYScale]{},
		Y2: &axisConfigAPIModel[kbapi.KibanaHTTPAPIsVisApiXyAxisConfig_Y2_Domain, kbapi.KibanaHTTPAPIsVisApiXyAxisConfigY2Scale]{},
	}

	m := &models.XYAxisModel{}
	diags := xyAxisFromAPI(m, apiAxis)
	require.False(t, diags.HasError(), "%v", diags)
	require.Nil(t, m.X)
	require.Nil(t, m.Y)
	require.Nil(t, m.Y2)
}

func TestAxisConfigIsEmpty(t *testing.T) {
	require.True(t, axisConfigIsEmpty(nil))
	require.True(t, axisConfigIsEmpty(&models.YAxisConfigModel{}))
	require.False(t, axisConfigIsEmpty(&models.YAxisConfigModel{Grid: types.BoolValue(true)}))
	require.False(t, axisConfigIsEmpty(&models.YAxisConfigModel{Scale: types.StringValue("log")}))
}
