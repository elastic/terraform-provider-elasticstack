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

type processorRerouteModel struct {
	CommonProcessorModel
	Destination types.String `tfsdk:"destination"`
	Dataset     types.String `tfsdk:"dataset"`
	Namespace   types.String `tfsdk:"namespace"`
}

func (m *processorRerouteModel) TypeName() string { return "reroute" }

func (m *processorRerouteModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorRerouteBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if IsKnown(m.Destination) {
		body.Destination = m.Destination.ValueString()
	}
	if IsKnown(m.Dataset) {
		body.Dataset = m.Dataset.ValueString()
	}
	if IsKnown(m.Namespace) {
		body.Namespace = m.Namespace.ValueString()
	}

	return body, diags
}

// NewProcessorRerouteDataSource returns a PF data source for the reroute processor.
func NewProcessorRerouteDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource.",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"destination": schema.StringAttribute{
			Description: "The destination data stream, index, or index alias to route the document to.",
			Optional:    true,
		},
		"dataset": schema.StringAttribute{
			Description: "The destination dataset to route the document to.",
			Optional:    true,
		},
		"namespace": schema.StringAttribute{
			Description: "The destination namespace to route the document to.",
			Optional:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorRerouteModel{}, schema.Schema{
		Description: processorRerouteDataSourceDescription,
		Attributes:  attrs,
	})
}
