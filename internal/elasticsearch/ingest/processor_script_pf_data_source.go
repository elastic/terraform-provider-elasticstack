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
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorScriptModel struct {
	CommonProcessorModel
	ID       types.String         `tfsdk:"id"`
	JSON     types.String         `tfsdk:"json"`
	Lang     types.String         `tfsdk:"lang"`
	ScriptID types.String         `tfsdk:"script_id"`
	Source   types.String         `tfsdk:"source"`
	Params   jsontypes.Normalized `tfsdk:"params"`
}

func (m *processorScriptModel) TypeName() string    { return "script" }
func (m *processorScriptModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorScriptModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorScriptModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorScriptBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Lang) {
		body.Lang = m.Lang.ValueString()
	}
	if IsKnown(m.ScriptID) {
		body.ScriptID = m.ScriptID.ValueString()
	}
	if IsKnown(m.Source) {
		body.Source = m.Source.ValueString()
	}
	if IsKnown(m.Params) {
		params := make(map[string]any)
		if err := json.Unmarshal([]byte(m.Params.ValueString()), &params); err != nil {
			diags.AddError("Failed to parse params JSON", err.Error())
			return nil, diags
		}
		body.Params = params
	}

	// Ensure ignore_failure default is reflected in state.
	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorScriptDataSource returns a PF data source for the script processor.
func NewProcessorScriptDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource.",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"lang": schema.StringAttribute{
			Description: "Script language.",
			Optional:    true,
		},
		"script_id": schema.StringAttribute{
			Description: "ID of a stored script. If no `source` is specified, this parameter is required.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("source"),
				),
			},
		},
		"source": schema.StringAttribute{
			Description: "Inline script. If no `script_id` is specified, this parameter is required.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("script_id"),
				),
			},
		},
		"params": schema.StringAttribute{
			Description: "Object containing parameters for the script.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorScriptModel{}, schema.Schema{
		Description: processorScriptDataSourceDescription,
		Attributes:  attrs,
	})
}
