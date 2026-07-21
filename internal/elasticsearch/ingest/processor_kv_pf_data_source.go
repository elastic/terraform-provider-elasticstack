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

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorKVModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	FieldSplit    types.String `tfsdk:"field_split"`
	ValueSplit    types.String `tfsdk:"value_split"`
	IncludeKeys   types.Set    `tfsdk:"include_keys"`
	ExcludeKeys   types.Set    `tfsdk:"exclude_keys"`
	Prefix        types.String `tfsdk:"prefix"`
	TrimKey       types.String `tfsdk:"trim_key"`
	TrimValue     types.String `tfsdk:"trim_value"`
	StripBrackets types.Bool   `tfsdk:"strip_brackets"`
}

func (m *processorKVModel) TypeName() string { return "kv" }

func (m *processorKVModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorKVBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody(false)

	if typeutils.IsKnown(m.FieldSplit) {
		body.FieldSplit = m.FieldSplit.ValueString()
	}
	if typeutils.IsKnown(m.ValueSplit) {
		body.ValueSplit = m.ValueSplit.ValueString()
	}
	body.IncludeKeys = typeutils.StringSetElements(m.IncludeKeys, &diags)
	body.ExcludeKeys = typeutils.StringSetElements(m.ExcludeKeys, &diags)
	if typeutils.IsKnown(m.Prefix) {
		body.Prefix = m.Prefix.ValueString()
	}
	if typeutils.IsKnown(m.TrimKey) {
		body.TrimKey = m.TrimKey.ValueString()
	}
	if typeutils.IsKnown(m.TrimValue) {
		body.TrimValue = m.TrimValue.ValueString()
	}
	if m.StripBrackets.IsNull() || m.StripBrackets.IsUnknown() {
		m.StripBrackets = types.BoolValue(false)
		body.StripBrackets = false
	} else {
		body.StripBrackets = m.StripBrackets.ValueBool()
	}

	return body, diags
}

// NewProcessorKVDataSource returns a PF data source for the kv processor.
func NewProcessorKVDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: descIdentifierWithPeriod,
			Computed:    true,
		},
		attrJSON: schema.StringAttribute{
			Description: descJSONDataSource,
			Computed:    true,
		},
		attrField: schema.StringAttribute{
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
		attrTargetField: schema.StringAttribute{
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
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
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
