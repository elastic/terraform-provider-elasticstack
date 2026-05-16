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

type processorDotExpanderModel struct {
	CommonProcessorModel
	Field    types.String `tfsdk:"field"`
	Path     types.String `tfsdk:"path"`
	Override types.Bool   `tfsdk:"override"`
}

func (m *processorDotExpanderModel) TypeName() string { return "dot_expander" }

func (m *processorDotExpanderModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorDotExpanderBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Field) {
		body.Field = m.Field.ValueString()
	}
	if IsKnown(m.Path) {
		body.Path = m.Path.ValueString()
	}
	if m.Override.IsNull() || m.Override.IsUnknown() {
		m.Override = types.BoolValue(false)
		body.Override = false
	} else {
		body.Override = m.Override.ValueBool()
	}

	return body, diags
}

// NewProcessorDotExpanderDataSource returns a PF data source for the dot_expander processor.
func NewProcessorDotExpanderDataSource() datasource.DataSource {
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
			Description: "The field to expand into an object field. If set to *, all top-level fields will be expanded.",
			Required:    true,
		},
		"path": schema.StringAttribute{
			Description: "The field that contains the field to expand.",
			Optional:    true,
		},
		"override": schema.BoolAttribute{
			Description: "Controls the behavior when there is already an existing nested object that conflicts with the expanded field.",
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorDotExpanderModel{}, schema.Schema{
		Description: processorDotExpanderDataSourceDescription,
		Attributes:  attrs,
	})
}
