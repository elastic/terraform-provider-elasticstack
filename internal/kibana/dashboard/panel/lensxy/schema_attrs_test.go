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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestXYMetricConfigJSONDocumentsDurationFormatMinimumVersion(t *testing.T) {
	t.Parallel()

	yAttr, ok := xyDataLayerAttributes()["y"].(schema.ListNestedAttribute)
	require.True(t, ok)
	configAttr, ok := yAttr.NestedObject.Attributes["config_json"].(schema.StringAttribute)
	require.True(t, ok)

	require.Contains(t, configAttr.MarkdownDescription, "9.5.2")
	require.Contains(t, configAttr.MarkdownDescription, "duration")
	require.Contains(t, configAttr.MarkdownDescription, "decimals")
	require.Contains(t, configAttr.MarkdownDescription, "compact")
}

func TestReferenceLineThresholdAxisValidator(t *testing.T) {
	t.Parallel()

	thresholdsAttr, ok := xyReferenceLineLayerAttributes()["thresholds"].(schema.ListNestedAttribute)
	require.True(t, ok)
	axisAttr, ok := thresholdsAttr.NestedObject.Attributes["axis"].(schema.StringAttribute)
	require.True(t, ok)
	require.False(t, axisAttr.Computed)
	require.True(t, axisAttr.Optional)

	validate := func(value types.String) validator.StringResponse {
		resp := validator.StringResponse{}
		for _, v := range axisAttr.Validators {
			v.ValidateString(context.Background(), validator.StringRequest{ConfigValue: value}, &resp)
		}
		return resp
	}

	for _, ok := range []string{"x", "y", "y2"} {
		require.False(t, validate(types.StringValue(ok)).Diagnostics.HasError(), ok)
	}
	for _, v := range []types.String{types.StringNull(), types.StringUnknown()} {
		require.False(t, validate(v).Diagnostics.HasError(), v.String())
	}
	for _, bad := range []string{"left", "right", "bottom", "Y", "", "y1"} {
		resp := validate(types.StringValue(bad))
		require.True(t, resp.Diagnostics.HasError(), bad)
		detail := resp.Diagnostics[0].Detail()
		require.Contains(t, detail, "x")
		require.Contains(t, detail, "y2")
	}
}
