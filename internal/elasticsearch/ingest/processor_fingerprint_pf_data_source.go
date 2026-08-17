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
	Fields        types.List   `tfsdk:"fields"`
	TargetField   types.String `tfsdk:"target_field"`
	IgnoreMissing types.Bool   `tfsdk:"ignore_missing"`
	Salt          types.String `tfsdk:"salt"`
	Method        types.String `tfsdk:"method"`
}

func (m *processorFingerprintModel) TypeName() string { return "fingerprint" }

func (m *processorFingerprintModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorFingerprintBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	body.Fields = typeutils.StringListElements(m.Fields, &diags)
	body.TargetField = typeutils.StringDefault(&m.TargetField, "fingerprint")
	body.IgnoreMissing = typeutils.BoolDefault(&m.IgnoreMissing, false)
	if typeutils.IsKnown(m.Salt) {
		body.Salt = m.Salt.ValueString()
	}
	body.Method = typeutils.StringDefault(&m.Method, "SHA-1")

	typeutils.BoolDefault(&m.IgnoreFailure, false)

	return body, diags
}

// NewProcessorFingerprintDataSource returns a PF data source for the fingerprint processor.
func NewProcessorFingerprintDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"fields": schema.ListAttribute{
			Description: "Array of fields to include in the fingerprint.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		attrTargetField: schema.StringAttribute{
			Description: "Output field for the fingerprint.",
			Optional:    true,
			Computed:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
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
