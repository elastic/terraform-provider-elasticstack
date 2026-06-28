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

package templateutil

import (
	"context"
	"encoding/json"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stringValue interface {
	IsNull() bool
	IsUnknown() bool
	ValueString() string
}

// IsKnownSemanticallyEmpty reports whether a prior JSON string value is a
// known, non-null value that nevertheless decodes to a zero-length JSON object
// (for example `{}` or whitespace-padded variants). The flatten layer uses this
// signal to preserve a practitioner-authored empty-object value in state when
// the Elasticsearch GET response omits the corresponding field entirely.
func IsKnownSemanticallyEmpty(v stringValue) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return typeutils.IsEmptyJSONObject(v.ValueString())
}

// FlattenTemplateCoreResult holds the shared template block fields produced by
// FlattenTemplateCore. Callers extend this with package-specific fields (e.g.
// lifecycle in the index template package) before constructing the final
// types.Object.
type FlattenTemplateCoreResult struct {
	AliasSet types.Set
	Mappings esindex.MappingsValue
	Settings customtypes.IndexSettingsValue
	DsoObj   types.Object
}

// FlattenTemplateCore extracts the shared template block fields from a
// *models.Template: aliases (sorted, with optional routing preservation),
// mappings, settings, and data_stream_options.
//
// priorMappings and priorSettings are consulted only when the API response
// contains no mappings/settings: when the prior value is a known,
// semantically-empty JSON object the prior value is preserved in state to
// avoid the post-apply consistency error the Plugin Framework raises when a
// planned "{}" collides with a flattened null.
//
// priorRouting carries user-configured alias routing values to restore when
// the API omits them. Pass nil to skip routing preservation.
//
// aliasElemType is the element type for the alias set (callers may supply a
// custom object type such as template.AliasObjectType).
// aliasAttrTypes is the attribute type map used to construct each alias element.
func FlattenTemplateCore(
	ctx context.Context,
	t *models.Template,
	priorMappings esindex.MappingsValue,
	priorSettings customtypes.IndexSettingsValue,
	priorRouting map[string]string,
	aliasElemType attr.Type,
	aliasAttrTypes map[string]attr.Type,
) (FlattenTemplateCoreResult, diag.Diagnostics) {
	var diags diag.Diagnostics
	var result FlattenTemplateCoreResult

	aliasSet, d := aliasutil.FlattenAliasSet(ctx, t.Aliases, priorRouting, aliasElemType, aliasAttrTypes)
	diags.Append(d...)
	if diags.HasError() {
		return result, diags
	}
	result.AliasSet = aliasSet

	mappings, mappingsDiags := flattenMappings(t.Mappings, priorMappings)
	diags.Append(mappingsDiags...)
	if diags.HasError() {
		return result, diags
	}
	result.Mappings = mappings

	settings, settingsDiags := flattenSettings(t.Settings, priorSettings)
	diags.Append(settingsDiags...)
	if diags.HasError() {
		return result, diags
	}
	result.Settings = settings

	var dsoObj types.Object
	if t.DataStreamOptions != nil && t.DataStreamOptions.FailureStore != nil {
		var dsoDiags diag.Diagnostics
		dsoObj, dsoDiags = datastreamoptions.FlattenLocal(t.DataStreamOptions)
		diags.Append(dsoDiags...)
		if diags.HasError() {
			return result, diags
		}
	} else {
		dsoObj = types.ObjectNull(datastreamoptions.AttrTypes())
	}
	result.DsoObj = dsoObj

	return result, diags
}

// flattenMappings maps the API mappings response onto a MappingsValue, applying
// the prior-preservation rule when the API returns no mappings.
func flattenMappings(apiMappings map[string]any, prior esindex.MappingsValue) (esindex.MappingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiMappings) > 0 {
		b, err := json.Marshal(apiMappings)
		if err != nil {
			diags.AddError("Failed to marshal template.mappings", err.Error())
			return esindex.NewMappingsNull(), diags
		}
		return esindex.NewMappingsValue(string(b)), diags
	}
	if IsKnownSemanticallyEmpty(prior) {
		return prior, diags
	}
	return esindex.NewMappingsNull(), diags
}

// flattenSettings maps the API settings response onto an IndexSettingsValue,
// applying the prior-preservation rule when the API returns no settings.
func flattenSettings(apiSettings map[string]any, prior customtypes.IndexSettingsValue) (customtypes.IndexSettingsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiSettings) > 0 {
		b, err := json.Marshal(apiSettings)
		if err != nil {
			diags.AddError("Failed to marshal template.settings", err.Error())
			return customtypes.NewIndexSettingsNull(), diags
		}
		return customtypes.NewIndexSettingsValue(string(b)), diags
	}
	if IsKnownSemanticallyEmpty(prior) {
		return prior, diags
	}
	return customtypes.NewIndexSettingsNull(), diags
}
