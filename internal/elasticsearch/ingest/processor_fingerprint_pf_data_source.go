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

type processorFingerprintModel struct {
	CommonProcessorModel
	ID            types.String `tfsdk:"id"`
	JSON          types.String `tfsdk:"json"`
	Fields        types.List   `tfsdk:"fields"`
	TargetField   types.String `tfsdk:"target_field"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
	Salt          types.String `tfsdk:"salt"`
	Method        types.String `tfsdk:"method"`
}

func (m *processorFingerprintModel) TypeName() string    { return "fingerprint" }
func (m *processorFingerprintModel) SetID(id string)     { m.ID = types.StringValue(id) }
func (m *processorFingerprintModel) SetJSON(json string) { m.JSON = types.StringValue(json) }

func (m *processorFingerprintModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorFingerprintBody{}

	commonBody, d := toCommonProcessorBody(m.CommonProcessorModel)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	body.CommonProcessorBody = commonBody

	if IsKnown(m.Fields) {
		elems := make([]string, 0, len(m.Fields.Elements()))
		for _, elem := range m.Fields.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid fields element type", "expected types.String")
				} else {
					diags.AddError("Unknown fields element", "fields elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.Fields = elems
	}
	if m.TargetField.IsNull() || m.TargetField.IsUnknown() {
		m.TargetField = types.StringValue("fingerprint")
		body.TargetField = "fingerprint"
	} else {
		body.TargetField = m.TargetField.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}
	if IsKnown(m.Salt) {
		body.Salt = m.Salt.ValueString()
	}
	if m.Method.IsNull() || m.Method.IsUnknown() {
		m.Method = types.StringValue("SHA-1")
		body.Method = "SHA-1"
	} else {
		body.Method = m.Method.ValueString()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorFingerprintDataSource returns a PF data source for the fingerprint processor.
func NewProcessorFingerprintDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"fields": schema.ListAttribute{
			Description: "Array of fields to include in the fingerprint.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"target_field": schema.StringAttribute{
			Description: "Output field for the fingerprint.",
			Optional:    true,
			Computed:    true,
		},
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true`, the processor ignores any missing `fields`. If all fields are missing, the processor silently exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"salt": schema.StringAttribute{
			Description: "Salt value for the hash function.",
			Optional:    true,
		},
		"method": schema.StringAttribute{
			Description: "The hash method used to compute the fingerprint.",
			Optional:    true,
			Computed:    true,
			Validators:  []validator.String{stringvalidator.OneOf("MD5", "SHA-1", "SHA-256", "SHA-512", "MurmurHash3")},
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorFingerprintModel{}, schema.Schema{
		Description: processorFingerprintDataSourceDescription,
		Attributes:  attrs,
	})
}
