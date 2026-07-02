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

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestApplyPresentationFromAPI_knownFieldsUpdated(t *testing.T) {
	t.Parallel()

	title := types.StringValue("old-title")
	desc := types.StringValue("old-desc")
	hideTitle := types.BoolValue(false)
	hideBorder := types.BoolValue(false)

	ApplyPresentationFromAPI(
		&title, &desc, &hideTitle, &hideBorder,
		strPtr("new-title"), strPtr("new-desc"), boolPtr(true), boolPtr(true),
	)

	assert.Equal(t, "new-title", title.ValueString())
	assert.Equal(t, "new-desc", desc.ValueString())
	assert.True(t, hideTitle.ValueBool())
	assert.True(t, hideBorder.ValueBool())
}

func TestApplyPresentationFromAPI_nullFieldsPreserved(t *testing.T) {
	t.Parallel()

	title := types.StringNull()
	desc := types.StringNull()
	hideTitle := types.BoolNull()
	hideBorder := types.BoolNull()

	ApplyPresentationFromAPI(
		&title, &desc, &hideTitle, &hideBorder,
		strPtr("api-title"), strPtr("api-desc"), boolPtr(true), boolPtr(true),
	)

	assert.True(t, title.IsNull(), "null title should be preserved")
	assert.True(t, desc.IsNull(), "null desc should be preserved")
	assert.True(t, hideTitle.IsNull(), "null hide_title should be preserved")
	assert.True(t, hideBorder.IsNull(), "null hide_border should be preserved")
}

func TestApplyPresentationFromAPI_apiNilSetsNull(t *testing.T) {
	t.Parallel()

	title := types.StringValue("had-title")
	desc := types.StringValue("had-desc")
	hideTitle := types.BoolValue(true)
	hideBorder := types.BoolValue(true)

	ApplyPresentationFromAPI(&title, &desc, &hideTitle, &hideBorder, nil, nil, nil, nil)

	assert.True(t, title.IsNull(), "nil API string should yield null")
	assert.True(t, desc.IsNull())
	assert.True(t, hideTitle.IsNull())
	assert.True(t, hideBorder.IsNull())
}

func TestNullPreservePresentationFromPrior_nullPriorResetsExisting(t *testing.T) {
	t.Parallel()

	existingTitle := types.StringValue("title")
	existingDesc := types.StringValue("desc")
	existingHideTitle := types.BoolValue(true)
	existingHideBorder := types.BoolValue(true)

	NullPreservePresentationFromPrior(
		types.StringNull(), types.StringNull(), types.BoolNull(), types.BoolNull(),
		&existingTitle, &existingDesc, &existingHideTitle, &existingHideBorder,
	)

	assert.True(t, existingTitle.IsNull())
	assert.True(t, existingDesc.IsNull())
	assert.True(t, existingHideTitle.IsNull())
	assert.True(t, existingHideBorder.IsNull())
}

func TestNullPreservePresentationFromPrior_knownPriorLeavesExistingUnchanged(t *testing.T) {
	t.Parallel()

	existingTitle := types.StringValue("title")
	existingDesc := types.StringValue("desc")
	existingHideTitle := types.BoolValue(true)
	existingHideBorder := types.BoolValue(false)

	NullPreservePresentationFromPrior(
		types.StringValue("prior-title"), types.StringValue("prior-desc"),
		types.BoolValue(false), types.BoolValue(true),
		&existingTitle, &existingDesc, &existingHideTitle, &existingHideBorder,
	)

	assert.Equal(t, "title", existingTitle.ValueString())
	assert.Equal(t, "desc", existingDesc.ValueString())
	assert.True(t, existingHideTitle.ValueBool())
	assert.False(t, existingHideBorder.ValueBool())
}

func TestBuildPresentationConfig_writesKnownFields(t *testing.T) {
	t.Parallel()

	var apiTitle *string
	var apiDesc *string
	var apiHideTitle *bool
	var apiHideBorder *bool

	BuildPresentationConfig(
		types.StringValue("my-title"),
		types.StringValue("my-desc"),
		types.BoolValue(true),
		types.BoolValue(false),
		&apiTitle, &apiDesc, &apiHideTitle, &apiHideBorder,
	)

	assert.Equal(t, "my-title", *apiTitle)
	assert.Equal(t, "my-desc", *apiDesc)
	assert.True(t, *apiHideTitle)
	assert.False(t, *apiHideBorder)
}

func TestBuildPresentationConfig_nullFieldsSkipped(t *testing.T) {
	t.Parallel()

	var apiTitle *string
	var apiDesc *string
	var apiHideTitle *bool
	var apiHideBorder *bool

	BuildPresentationConfig(
		types.StringNull(), types.StringNull(), types.BoolNull(), types.BoolNull(),
		&apiTitle, &apiDesc, &apiHideTitle, &apiHideBorder,
	)

	assert.Nil(t, apiTitle, "null cfg should not set API title")
	assert.Nil(t, apiDesc)
	assert.Nil(t, apiHideTitle)
	assert.Nil(t, apiHideBorder)
}
