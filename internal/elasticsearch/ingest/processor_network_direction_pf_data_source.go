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

	if IsKnown(m.SourceIP) {
		body.SourceIP = m.SourceIP.ValueString()
	}
	if IsKnown(m.DestinationIP) {
		body.DestinationIP = m.DestinationIP.ValueString()
	}
	if IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if IsKnown(m.InternalNetworks) {
		elems := make([]string, 0, len(m.InternalNetworks.Elements()))
		for _, elem := range m.InternalNetworks.Elements() {
			str, ok := elem.(types.String)
			if !ok || !IsKnown(str) {
				if !ok {
					diags.AddError("Invalid internal_networks element type", "expected types.String")
				} else {
					diags.AddError("Unknown internal_networks element", "internal_networks elements cannot be unknown")
				}
				continue
			}
			elems = append(elems, str.ValueString())
		}
		body.InternalNetworks = elems
	}
	if IsKnown(m.InternalNetworksField) {
		body.InternalNetworksField = m.InternalNetworksField.ValueString()
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(true)
		body.IgnoreMissing = true
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorNetworkDirectionDataSource returns a PF data source for the network_direction processor.
func NewProcessorNetworkDirectionDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource.",
			Computed:    true,
		},
		"json": schema.StringAttribute{
			Description: "JSON representation of this data source.",
			Computed:    true,
		},
		"source_ip": schema.StringAttribute{
			Description: "Field containing the source IP address.",
			Optional:    true,
		},
		"destination_ip": schema.StringAttribute{
			Description: "Field containing the destination IP address.",
			Optional:    true,
		},
		"target_field": schema.StringAttribute{
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
		"ignore_missing": schema.BoolAttribute{
			Description: "If `true` and `field` does not exist or is `null`, the processor quietly exits without modifying the document.",
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
