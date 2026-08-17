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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorAppendModel struct {
	CommonProcessorModel
	Field           types.String `tfsdk:"field"`
	Value           types.List   `tfsdk:"value"`
	AllowDuplicates types.Bool   `tfsdk:"allow_duplicates"`
	MediaType       types.String `tfsdk:"media_type"`
}

func (m *processorAppendModel) TypeName() string { return processorTypeAppend }

func (m *processorAppendModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorAppendBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}

	body.Value = typeutils.StringListElements(m.Value, &diags)

	body.AllowDuplicates = typeutils.BoolDefault(&m.AllowDuplicates, true)

	if typeutils.IsKnown(m.MediaType) {
		body.MediaType = m.MediaType.ValueString()
	}

	return body, diags
}

// NewProcessorAppendDataSource returns a PF data source for the append processor.
func NewProcessorAppendDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		attrField: schema.StringAttribute{
			Description: "The field to be appended to.",
			Required:    true,
		},
		attrValue: schema.ListAttribute{
			Description: "The value to be appended.",
			Required:    true,
			ElementType: types.StringType,
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		attrAllowDuplicates: schema.BoolAttribute{
			Description: "If `false`, the processor does not append values already present in the field.",
			Optional:    true,
			Computed:    true,
		},
		"media_type": schema.StringAttribute{
			Description: processorAppendMediaTypeDescription,
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorAppendModel{}, schema.Schema{
		Description: processorAppendDataSourceDescription,
		Attributes:  attrs,
	})
}
