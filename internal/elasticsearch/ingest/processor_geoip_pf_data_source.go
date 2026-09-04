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

type processorGeoIPModel struct {
	CommonProcessorModel
	WithIgnorableTargetField
	DatabaseFile types.String `tfsdk:"database_file"`
	Properties   types.Set    `tfsdk:"properties"`
	FirstOnly    types.Bool   `tfsdk:"first_only"`
}

func (m *processorGeoIPModel) TypeName() string { return "geoip" }

func (m *processorGeoIPModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorGeoIPBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}
	body.WithIgnorableTargetFieldBody = m.toIgnorableTargetFieldBody()

	body.TargetField = typeutils.StringDefault(&m.TargetField, "geoip")
	if typeutils.IsKnown(m.DatabaseFile) {
		body.DatabaseFile = m.DatabaseFile.ValueString()
	}
	body.Properties = typeutils.StringElements(m.Properties, &diags)
	body.FirstOnly = typeutils.BoolDefault(&m.FirstOnly, true)

	return body, diags
}

// NewProcessorGeoIPDataSource returns a PF data source for the geoip processor.
func NewProcessorGeoIPDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		attrField: schema.StringAttribute{
			Description: "The field to get the ip address from for the geographical lookup.",
			Required:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: "The field that will hold the geographical information looked up from the MaxMind database.",
			Optional:    true,
			Computed:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: "If `true` and `field` does not exist, the processor quietly exits without modifying the document.",
			Optional:    true,
			Computed:    true,
		},
		"database_file": schema.StringAttribute{
			Description: processorGeoIPDatabaseFileDescription,
			Optional:    true,
		},
		attrProperties: schema.SetAttribute{
			Description: "Controls what properties are added to the `target_field` based on the geoip lookup.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"first_only": schema.BoolAttribute{
			Description: "If `true` only first found geoip data will be returned, even if field contains array.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorGeoIPModel{}, schema.Schema{
		Description: processorGeoIPDataSourceDescription,
		Attributes:  attrs,
	})
}
