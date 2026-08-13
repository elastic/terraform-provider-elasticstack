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

package fleetschema

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/stretchr/testify/require"
)

func defaultBoolValue(t *testing.T, d defaults.Bool) bool {
	t.Helper()
	var resp defaults.BoolResponse
	d.DefaultBool(context.Background(), defaults.BoolRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp.PlanValue.ValueBool()
}

func TestBoolAttributeBuilders(t *testing.T) {
	builders := map[string]func(computed bool, def *bool) schema.BoolAttribute{
		"IgnoreMappingUpdateErrorsAttribute": IgnoreMappingUpdateErrorsAttribute,
		"SkipDataStreamRolloverAttribute":    SkipDataStreamRolloverAttribute,
		"SkipDestroyAttribute":               SkipDestroyAttribute,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			t.Run("optional only", func(t *testing.T) {
				attr := build(false, nil)
				require.True(t, attr.Optional)
				require.False(t, attr.Computed)
				require.Nil(t, attr.Default)
				require.NotEmpty(t, attr.Description)
			})

			t.Run("computed with default", func(t *testing.T) {
				attr := build(true, new(bool))
				require.True(t, attr.Optional)
				require.True(t, attr.Computed)
				require.NotNil(t, attr.Default)
				require.False(t, defaultBoolValue(t, attr.Default))
			})
		})
	}
}

func TestBoolAttributeBuildersShareDescriptions(t *testing.T) {
	require.Equal(t,
		IgnoreMappingUpdateErrorsAttribute(false, nil).Description,
		IgnoreMappingUpdateErrorsAttribute(true, new(bool)).Description,
	)
	require.Equal(t,
		SkipDataStreamRolloverAttribute(false, nil).Description,
		SkipDataStreamRolloverAttribute(true, new(bool)).Description,
	)
	require.Equal(t,
		SkipDestroyAttribute(false, nil).Description,
		SkipDestroyAttribute(true, new(bool)).Description,
	)
}
