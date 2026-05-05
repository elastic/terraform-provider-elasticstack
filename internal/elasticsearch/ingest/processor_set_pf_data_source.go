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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorSetModel struct {
	CommonProcessorModel
	Field            types.String `tfsdk:"field"`
	Value            types.String `tfsdk:"value"`
	CopyFrom         types.String `tfsdk:"copy_from"`
	Override         types.Bool   `tfsdk:"override"`
	IgnoreEmptyValue types.Bool   `tfsdk:"ignore_empty_value"`
	MediaType        types.String `tfsdk:"media_type"`
}

func (m *processorSetModel) TypeName() string { return "set" }

func (m *processorSetModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorSetBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.Value) {
		body.Value = m.Value.ValueString()
	}
	if IsKnown(m.CopyFrom) {
		body.CopyFrom = m.CopyFrom.ValueString()
	}
	if m.Override.IsNull() || m.Override.IsUnknown() {
		m.Override = types.BoolValue(true)
		body.Override = true
	} else {
		body.Override = m.Override.ValueBool()
	}
	if m.IgnoreEmptyValue.IsNull() || m.IgnoreEmptyValue.IsUnknown() {
		m.IgnoreEmptyValue = types.BoolValue(false)
		body.IgnoreEmptyValue = false
	} else {
		body.IgnoreEmptyValue = m.IgnoreEmptyValue.ValueBool()
	}
	if m.MediaType.IsNull() || m.MediaType.IsUnknown() {
		m.MediaType = types.StringValue("application/json")
		body.MediaType = "application/json"
	} else {
		body.MediaType = m.MediaType.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorSetDataSource returns a PF data source for the set processor.
func NewProcessorSetDataSource() datasource.DataSource {
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
			Description: "The field to insert, upsert, or update.",
			Required:    true,
		},
		"value": schema.StringAttribute{
			Description: "The value to be set for the field. Supports template snippets. May specify only one of `value` or `copy_from`.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("copy_from"),
				),
			},
		},
		"copy_from": schema.StringAttribute{
			Description: "The origin field which will be copied to `field`, cannot set `value` simultaneously.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.ExactlyOneOf(
					path.MatchRelative().AtParent().AtName("value"),
				),
			},
		},
		"override": schema.BoolAttribute{
			Description: "If processor will update fields with pre-existing non-null-valued field.",
			Optional:    true,
			Computed:    true,
		},
		"ignore_empty_value": schema.BoolAttribute{
			Description: "If `true` and `value` is a template snippet that evaluates to `null` or the empty string, the processor quietly exits without modifying the document",
			Optional:    true,
			Computed:    true,
		},
		"media_type": schema.StringAttribute{
			Description: "The media type for encoding value.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorSetModel{}, schema.Schema{
		Description: processorSetDataSourceDescription,
		Attributes:  attrs,
	})
}
