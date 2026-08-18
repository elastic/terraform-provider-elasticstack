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

type processorNetworkDirectionModel struct {
	CommonProcessorModel
	SourceIP              types.String `tfsdk:"source_ip"`
	DestinationIP         types.String `tfsdk:"destination_ip"`
	TargetField           types.String `tfsdk:"target_field"`
	InternalNetworks      types.Set    `tfsdk:"internal_networks"`
	InternalNetworksField types.String `tfsdk:"internal_networks_field"`
	IgnoreMissing         types.Bool   `tfsdk:"ignore_missing"`
}

func (m *processorNetworkDirectionModel) TypeName() string { return "network_direction" }

func (m *processorNetworkDirectionModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorNetworkDirectionBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.SourceIP) {
		body.SourceIP = m.SourceIP.ValueString()
	}
	if typeutils.IsKnown(m.DestinationIP) {
		body.DestinationIP = m.DestinationIP.ValueString()
	}
	if typeutils.IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	body.InternalNetworks = typeutils.StringSetElements(m.InternalNetworks, &diags)
	if typeutils.IsKnown(m.InternalNetworksField) {
		body.InternalNetworksField = m.InternalNetworksField.ValueString()
	}
	body.IgnoreMissing = typeutils.BoolDefault(&m.IgnoreMissing, true)

	typeutils.BoolDefault(&m.IgnoreFailure, false)

	return body, diags
}

// NewProcessorNetworkDirectionDataSource returns a PF data source for the network_direction processor.
func NewProcessorNetworkDirectionDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"source_ip": schema.StringAttribute{
			Description: "Field containing the source IP address.",
			Optional:    true,
		},
		"destination_ip": schema.StringAttribute{
			Description: "Field containing the destination IP address.",
			Optional:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: "Output field for the network direction.",
			Optional:    true,
		},
		"internal_networks": schema.SetAttribute{
			Description: "List of internal networks.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"internal_networks_field": schema.StringAttribute{
			Description: "A field on the given document to read the internal_networks configuration from.",
			Optional:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorNetworkDirectionModel{}, schema.Schema{
		Description: processorNetworkDirectionDataSourceDescription,
		Attributes:  attrs,
	})
}
