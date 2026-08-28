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

type processorSetSecurityUserModel struct {
	CommonProcessorModel
	Field      types.String `tfsdk:"field"`
	Properties types.Set    `tfsdk:"properties"`
}

func (m *processorSetSecurityUserModel) TypeName() string { return "set_security_user" }

func (m *processorSetSecurityUserModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorSetSecurityUserBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	body.Properties = typeutils.StringElements(m.Properties, &diags)

	return body, diags
}

// NewProcessorSetSecurityUserDataSource returns a PF data source for the set_security_user processor.
func NewProcessorSetSecurityUserDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		attrField: schema.StringAttribute{
			Description: "The field to store the user information into.",
			Required:    true,
		},
		attrProperties: schema.SetAttribute{
			Description: "Controls what user related properties are added to the `field`.",
			Optional:    true,
			ElementType: types.StringType,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorSetSecurityUserModel{}, schema.Schema{
		Description: processorSetSecurityUserDataSourceDescription,
		Attributes:  attrs,
	})
}
