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

package ingest

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CommonProcessorModel holds the fields shared by every ingest processor data
// source. Embed this struct in concrete processor models.
type CommonProcessorModel struct {
	Description   types.String `tfsdk:"description"`
	If            types.String `tfsdk:"if"`
	IgnoreFailure types.Bool   `tfsdk:"ignore_failure"`
	OnFailure     types.List   `tfsdk:"on_failure"`
	Tag           types.String `tfsdk:"tag"`
}

// CommonProcessorSchemaAttributes returns the schema attributes that are common
// to all ingest processor data sources.
func CommonProcessorSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Description of the processor.",
			Optional:    true,
		},
		"if": schema.StringAttribute{
			Description: "Conditionally execute the processor",
			Optional:    true,
		},
		"ignore_failure": schema.BoolAttribute{
			Description: "Ignore failures for the processor.",
			Optional:    true,
			Computed:    true,
		},
		"on_failure": schema.ListAttribute{
			Description: "Handle failures for the processor.",
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
		},
		"tag": schema.StringAttribute{
			Description: "Identifier for the processor.",
			Optional:    true,
		},
	}
}

// appendCommonFields populates dst with the common processor fields from model.
// It returns any diagnostics collected while parsing on_failure JSON values.
//
//nolint:unused // will be used by concrete processor MarshalBody implementations
func appendCommonFields(dst map[string]any, model CommonProcessorModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if IsKnown(model.Description) {
		dst["description"] = model.Description.ValueString()
	}
	if IsKnown(model.If) {
		dst["if"] = model.If.ValueString()
	}
	if IsKnown(model.IgnoreFailure) {
		dst["ignore_failure"] = model.IgnoreFailure.ValueBool()
	}
	if IsKnown(model.OnFailure) {
		elems := make([]map[string]any, 0, len(model.OnFailure.Elements()))
		for _, elem := range model.OnFailure.Elements() {
			norm, ok := elem.(jsontypes.Normalized)
			if !ok {
				diags.AddError("Invalid on_failure element type", "expected jsontypes.Normalized")
				continue
			}
			if !IsKnown(norm) {
				diags.AddError("Unknown on_failure element", "on_failure elements cannot be unknown")
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(norm.ValueString()), &item); err != nil {
				diags.AddError("Failed to parse on_failure JSON", err.Error())
				continue
			}
			elems = append(elems, item)
		}
		if len(elems) > 0 {
			dst["on_failure"] = elems
		}
	}
	if IsKnown(model.Tag) {
		dst["tag"] = model.Tag.ValueString()
	}

	return diags
}
