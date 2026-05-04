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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorKVModel struct {
	CommonProcessorModel
	ID            types.String `tfsdk:"id"`
	JSON          types.String `tfsdk:"json"`
	Field         types.String `tfsdk:"field"`
	TargetField   types.String `tfsdk:"target_field"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
	FieldSplit    types.String `tfsdk:"field_split"`
	ValueSplit    types.String `tfsdk:"value_split"`
	IncludeKeys   types.Set    `tfsdk:"include_keys"`
	ExcludeKeys   types.Set    `tfsdk:"exclude_keys"`
	Prefix        types.String `tfsdk:"prefix"`
	TrimKey       types.String `tfsdk:"trim_key"`
	TrimValue     types.String `tfsdk:"trim_value"`
	StripBrackets types.Bool   `tfsdk:"strip_brackets"`
}

func (m *processorKVModel) TypeName() string    { return "kv" }
func (m *processorKVModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorKVModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func stringSetValue(set types.Set, diags *diag.Diagnostics) []string {
	if !IsKnown(set) {
		return nil
	}
	elems := make([]string, 0, len(set.Elements()))
	for _, elem := range set.Elements() {
		str, ok := elem.(types.String)
		if !ok || !IsKnown(str) {
			if !ok {
				diags.AddError("Invalid list element type", "expected types.String")
			} else {
				diags.AddError("Unknown list element", "list elements cannot be unknown")
			}
			continue
		}
		elems = append(elems, str.ValueString())
	}
	return elems
}

func (m *processorKVModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorKVBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}
	if IsKnown(m.FieldSplit) {
		body.FieldSplit = m.FieldSplit.ValueString()
	}
	if IsKnown(m.ValueSplit) {
		body.ValueSplit = m.ValueSplit.ValueString()
	}
	body.IncludeKeys = stringSetValue(m.IncludeKeys, &diags)
	body.ExcludeKeys = stringSetValue(m.ExcludeKeys, &diags)
	if IsKnown(m.Prefix) {
		body.Prefix = m.Prefix.ValueString()
	}
	if IsKnown(m.TrimKey) {
		body.TrimKey = m.TrimKey.ValueString()
	}
	if IsKnown(m.TrimValue) {
		body.TrimValue = m.TrimValue.ValueString()
	}
	if m.StripBrackets.IsNull() || m.StripBrackets.IsUnknown() {
		m.StripBrackets = types.BoolValue(false)
		body.StripBrackets = false
	} else {
		body.StripBrackets = m.StripBrackets.ValueBool()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorKVDataSource returns a PF data source for the kv processor.
func NewProcessorKVDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"field": schema.StringAttribute{
			Description: "The field to be parsed. Supports template snippets.",
			Required:    true,
		},
		"field_split": schema.StringAttribute{
			Description: "Regex pattern to use for splitting key-value pairs.",
			Required:    true,
		},
		"value_split": schema.StringAttribute{
			Description: "Regex pattern to use for splitting the key from the value within a key-value pair.",
			Required:    true,
		},
		"target_field": schema.StringAttribute{
			Description: "The field to insert the extracted keys into. Defaults to the root of the document.",
			Optional:    true,
		},
		"include_keys": schema.SetAttribute{
			Description: "List of keys to filter and insert into document. Defaults to including all keys",
			Optional:    true,
			ElementType: types.StringType,
		},
		"exclude_keys": schema.SetAttribute{
			Description: "List of keys to exclude from document",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"prefix": schema.StringAttribute{
			Description: "Prefix to be added to extracted keys.",
			Optional:    true,
		},
		"trim_key": schema.StringAttribute{
			Description: "String of characters to trim from extracted keys.",
			Optional:    true,
		},
		"trim_value": schema.StringAttribute{
			Description: "String of characters to trim from extracted values.",
			Optional:    true,
		},
		"strip_brackets": schema.BoolAttribute{
			Description: "If `true` strip brackets `()`, `<>`, `[]` as well as quotes `'` and `\"` from extracted values.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorKVModel{}, schema.Schema{
		Description: processorKVDataSourceDescription,
		Attributes:  attrs,
	})
}
