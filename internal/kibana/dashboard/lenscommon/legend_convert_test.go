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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestLegendSizeTruncateVisibilityFromAPI_AllNil(t *testing.T) {
	t.Parallel()

	size, truncateAfterLines, visibility := LegendSizeTruncateVisibilityFromAPI(nil, nil, nil)

	assert.Equal(t, types.StringNull(), size)
	assert.Equal(t, types.Int64Null(), truncateAfterLines)
	assert.Equal(t, types.StringNull(), visibility)
}

func TestLegendSizeTruncateVisibilityFromAPI_FullyPopulated(t *testing.T) {
	t.Parallel()

	apiSize := "l"
	apiTruncate := float32(3)
	apiVisibility := "visible"

	size, truncateAfterLines, visibility := LegendSizeTruncateVisibilityFromAPI(&apiSize, &apiTruncate, &apiVisibility)

	assert.Equal(t, types.StringValue("l"), size)
	assert.Equal(t, types.Int64Value(3), truncateAfterLines)
	assert.Equal(t, types.StringValue("visible"), visibility)
}

func TestLegendSizeTruncateVisibilityToAPI_KnownFields(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	size, truncateAfterLines, visibility := LegendSizeTruncateVisibilityToAPI(
		types.StringValue("m"), types.StringValue("hidden"), types.Int64Value(5),
		"Missing legend size", "size must be provided",
		&diags,
	)

	assert.False(t, diags.HasError())
	if assert.NotNil(t, size) {
		assert.Equal(t, "m", *size)
	}
	if assert.NotNil(t, truncateAfterLines) {
		assert.InDelta(t, float32(5), *truncateAfterLines, 0)
	}
	if assert.NotNil(t, visibility) {
		assert.Equal(t, "hidden", *visibility)
	}
}

func TestLegendSizeTruncateVisibilityToAPI_UnknownOptionalFields(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	size, truncateAfterLines, visibility := LegendSizeTruncateVisibilityToAPI(
		types.StringValue("auto"), types.StringUnknown(), types.Int64Unknown(),
		"Missing legend size", "size must be provided",
		&diags,
	)

	assert.False(t, diags.HasError())
	assert.Nil(t, truncateAfterLines)
	assert.Nil(t, visibility)
	if assert.NotNil(t, size) {
		assert.Equal(t, "auto", *size)
	}
}

func TestLegendSizeTruncateVisibilityToAPI_MissingSize(t *testing.T) {
	t.Parallel()

	var diags diag.Diagnostics
	size, _, _ := LegendSizeTruncateVisibilityToAPI(
		types.StringNull(), types.StringNull(), types.Int64Null(),
		"Missing legend size", "size must be provided",
		&diags,
	)

	assert.Nil(t, size)
	assert.True(t, diags.HasError())
	assert.Equal(t, "Missing legend size", diags[0].Summary())
	assert.Equal(t, "size must be provided", diags[0].Detail())
}
