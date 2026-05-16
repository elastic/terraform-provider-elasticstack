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

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorGrokModel struct {
	CommonProcessorModel
	Field              types.String `tfsdk:"field"`
	Patterns           types.List   `tfsdk:"patterns"`
	PatternDefinitions types.Map    `tfsdk:"pattern_definitions"`
	EcsCompatibility   types.String `tfsdk:"ecs_compatibility"`
	TraceMatch         types.Bool   `tfsdk:"trace_match"`
	IgnoreMissing      types.Bool   `tfsdk:"ignore_missing"`
}

func (m *processorGrokModel) TypeName() string { return "grok" }

func (m *processorGrokModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorGrokBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.Patterns) {
		elems := make([]string, 0, len(m.Patterns.Elements()))
		for _, elem := range m.Patterns.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid patterns element type", "expected types.String")
				} else {
					diags.AddError("Unknown patterns element", "patterns elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.Patterns = elems
	}
	if IsKnown(m.PatternDefinitions) {
		defs := make(map[string]string, len(m.PatternDefinitions.Elements()))
		for k, v := range m.PatternDefinitions.Elements() {
			str, ok := v.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid pattern_definitions element type", "expected types.String")
				} else {
					diags.AddError("Unknown pattern_definitions element", "pattern_definitions elements cannot be unknown")
				}
				continue
			}
			defs[k] = str.ValueString()
		}
		body.PatternDefinitions = defs
	}
	if IsKnown(m.EcsCompatibility) {
		body.EcsCompatibility = m.EcsCompatibility.ValueString()
	}
	if m.TraceMatch.IsNull() || m.TraceMatch.IsUnknown() {
		m.TraceMatch = types.BoolValue(false)
		body.TraceMatch = false
	} else {
		body.TraceMatch = m.TraceMatch.ValueBool()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorGrokDataSource returns a PF data source for the grok processor.
func NewProcessorGrokDataSource() datasource.DataSource {
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
			Description: "The field to use for grok expression parsing",
			Required:    true,
		},
		"patterns": schema.ListAttribute{
			Description: "An ordered list of grok expression to match and extract named captures with. Returns on the first expression in the list that matches.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"pattern_definitions": schema.MapAttribute{
			Description: "A map of pattern-name and pattern tuples defining custom patterns to be used by the current processor. Patterns matching existing names will override the pre-existing definition.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"ecs_compatibility": schema.StringAttribute{
			Description: "Must be disabled or v1. If v1, the processor uses patterns with Elastic Common Schema (ECS) field names.",
			Optional:    true,
			Validators:  []validator.String{stringvalidator.OneOf("disabled", "v1")},
		},
		"trace_match": schema.BoolAttribute{
			Description: "when true, `_ingest._grok_match_index` will be inserted into your matched document's metadata with the index into the pattern found in `patterns` that matched.",
			Optional:    true,
			Computed:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorGrokModel{}, schema.Schema{
		Description: processorGrokDataSourceDescription,
		Attributes:  attrs,
	})
}
