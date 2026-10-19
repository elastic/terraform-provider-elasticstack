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

package panelkit

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// CopyKnownString

func TestCopyKnownString_known_returnsPointer(t *testing.T) {
	t.Parallel()
	got := CopyKnownString(types.StringValue("hello"))
	if assert.NotNil(t, got) {
		assert.Equal(t, "hello", *got)
	}
}

func TestCopyKnownString_null_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownString(types.StringNull()))
}

func TestCopyKnownString_unknown_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownString(types.StringUnknown()))
}

// CopyKnownBool

func TestCopyKnownBool_known_returnsPointer(t *testing.T) {
	t.Parallel()
	got := CopyKnownBool(types.BoolValue(true))
	if assert.NotNil(t, got) {
		assert.True(t, *got)
	}
}

func TestCopyKnownBool_null_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownBool(types.BoolNull()))
}

func TestCopyKnownBool_unknown_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownBool(types.BoolUnknown()))
}

// CopyKnownFloat32

func TestCopyKnownFloat32_known_returnsPointer(t *testing.T) {
	t.Parallel()
	got := CopyKnownFloat32(types.Float32Value(2.5))
	if assert.NotNil(t, got) {
		assert.InDelta(t, 2.5, *got, 1e-6)
	}
}

func TestCopyKnownFloat32_null_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownFloat32(types.Float32Null()))
}

func TestCopyKnownFloat32_unknown_returnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, CopyKnownFloat32(types.Float32Unknown()))
}

// BuildCommonControlFields

func TestBuildCommonControlFields_allKnown_copiesEveryField(t *testing.T) {
	t.Parallel()
	got := BuildCommonControlFields(CommonControlFields{
		Title:             types.StringValue("title"),
		UseGlobalFilters:  types.BoolValue(true),
		IgnoreValidations: types.BoolValue(false),
	})
	if assert.NotNil(t, got.Title) {
		assert.Equal(t, "title", *got.Title)
	}
	if assert.NotNil(t, got.UseGlobalFilters) {
		assert.True(t, *got.UseGlobalFilters)
	}
	if assert.NotNil(t, got.IgnoreValidations) {
		assert.False(t, *got.IgnoreValidations)
	}
}

func TestBuildCommonControlFields_allNullOrUnknown_leavesEveryFieldNil(t *testing.T) {
	t.Parallel()
	got := BuildCommonControlFields(CommonControlFields{
		Title:             types.StringNull(),
		UseGlobalFilters:  types.BoolUnknown(),
		IgnoreValidations: types.BoolNull(),
	})
	assert.Nil(t, got.Title)
	assert.Nil(t, got.UseGlobalFilters)
	assert.Nil(t, got.IgnoreValidations)
}
