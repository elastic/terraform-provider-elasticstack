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
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorJSONModel struct {
	CommonProcessorModel
	Field                     types.String `tfsdk:"field"`
	TargetField               types.String `tfsdk:"target_field"`
	AddToRoot                 types.Bool   `tfsdk:"add_to_root"`
	AddToRootConflictStrategy types.String `tfsdk:"add_to_root_conflict_strategy"`
	AllowDuplicateKeys        types.Bool   `tfsdk:"allow_duplicate_keys"`
}

func (m *processorJSONModel) TypeName() string { return "json" }

func (m *processorJSONModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorJSONBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if IsKnown(m.AddToRoot) {
		v := m.AddToRoot.ValueBool()
		body.AddToRoot = &v
	}
	if IsKnown(m.AddToRootConflictStrategy) {
		body.AddToRootConflictStrategy = m.AddToRootConflictStrategy.ValueString()
	}
	if IsKnown(m.AllowDuplicateKeys) {
		v := m.AllowDuplicateKeys.ValueBool()
		body.AllowDuplicateKeys = &v
	}

	return body, diags
}

// NewProcessorJSONDataSource returns a PF data source for the json processor.
func NewProcessorJSONDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource.",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"field": schema.StringAttribute{
			Description: "The field to be parsed.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field that the converted structured object will be written into. Any existing content in this field will be overwritten.",
			Optional:    true,
		},
		"add_to_root": schema.BoolAttribute{
			Description: "Flag that forces the parsed JSON to be added at the top level of the document. `target_field` must not be set when this option is chosen.",
			Optional:    true,
		},
		"add_to_root_conflict_strategy": schema.StringAttribute{
			Description: processorJSONAddToRootConflictDescription,
			Optional:    true,
			Validators:  []validator.String{stringvalidator.OneOf("replace", "merge")},
		},
		"allow_duplicate_keys": schema.BoolAttribute{
			Description: "When set to `true`, the JSON parser will not fail if the JSON contains duplicate keys. Instead, the last encountered value for any duplicate key wins.",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorJSONModel{}, schema.Schema{
		Description: processorJSONDataSourceDescription,
		Attributes:  attrs,
	})
}
