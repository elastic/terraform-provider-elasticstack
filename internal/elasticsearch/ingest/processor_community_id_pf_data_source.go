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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type processorCommunityIDModel struct {
	CommonProcessorModel
	SourceIP        types.String `tfsdk:"source_ip"`
	SourcePort      types.Int64  `tfsdk:"source_port"`
	DestinationIP   types.String `tfsdk:"destination_ip"`
	DestinationPort types.Int64  `tfsdk:"destination_port"`
	IanaNumber      types.Int64  `tfsdk:"iana_number"`
	IcmpType        types.Int64  `tfsdk:"icmp_type"`
	IcmpCode        types.Int64  `tfsdk:"icmp_code"`
	Seed            types.Int64  `tfsdk:"seed"`
	Transport       types.String `tfsdk:"transport"`
	TargetField     types.String `tfsdk:"target_field"`
	IgnoreMissing   types.Bool   `tfsdk:"ignore_missing"`
}

func (m *processorCommunityIDModel) TypeName() string { return "community_id" }

func (m *processorCommunityIDModel) MarshalBody() (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := processorCommunityIDBody{}

	body.CommonProcessorBody, diags = m.toCommonProcessorBody()
	if diags.HasError() {
		return nil, diags
	}

	if typeutils.IsKnown(m.SourceIP) {
		body.SourceIP = m.SourceIP.ValueString()
	}
	if typeutils.IsKnown(m.SourcePort) {
		body.SourcePort = typeutils.OptionalInt(m.SourcePort)
	}
	if typeutils.IsKnown(m.DestinationIP) {
		body.DestinationIP = m.DestinationIP.ValueString()
	}
	if typeutils.IsKnown(m.DestinationPort) {
		body.DestinationPort = typeutils.OptionalInt(m.DestinationPort)
	}
	if typeutils.IsKnown(m.IanaNumber) {
		body.IanaNumber = typeutils.OptionalInt(m.IanaNumber)
	}
	if typeutils.IsKnown(m.IcmpType) {
		body.IcmpType = typeutils.OptionalInt(m.IcmpType)
	}
	if typeutils.IsKnown(m.IcmpCode) {
		body.IcmpCode = typeutils.OptionalInt(m.IcmpCode)
	}
	if typeutils.IsKnown(m.Transport) {
		body.Transport = m.Transport.ValueString()
	}
	if typeutils.IsKnown(m.TargetField) {
		body.TargetField = m.TargetField.ValueString()
	}
	if m.Seed.IsNull() || m.Seed.IsUnknown() {
		m.Seed = types.Int64Value(0)
		body.Seed = new(0)
	} else {
		body.Seed = typeutils.OptionalInt(m.Seed)
	}
	if m.IgnoreMissing.IsNull() || m.IgnoreMissing.IsUnknown() {
		m.IgnoreMissing = types.BoolValue(false)
		body.IgnoreMissing = false
	} else {
		body.IgnoreMissing = m.IgnoreMissing.ValueBool()
	}

	if m.IgnoreFailure.IsNull() || m.IgnoreFailure.IsUnknown() {
		m.IgnoreFailure = types.BoolValue(false)
	}

	return body, diags
}

// NewProcessorCommunityIDDataSource returns a PF data source for the community_id processor.
func NewProcessorCommunityIDDataSource() datasource.DataSource {
	attrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: descIdentifierWithPeriod,
			Computed:    true,
		},
		attrJSON: schema.StringAttribute{
			Description: descJSONDataSource,
			Computed:    true,
		},
		"source_ip": schema.StringAttribute{
			Description: "Field containing the source IP address.",
			Optional:    true,
		},
		"source_port": schema.Int64Attribute{
			Description: "Field containing the source port.",
			Optional:    true,
		},
		"destination_ip": schema.StringAttribute{
			Description: "Field containing the destination IP address.",
			Optional:    true,
		},
		"destination_port": schema.Int64Attribute{
			Description: "Field containing the destination port.",
			Optional:    true,
		},
		"iana_number": schema.Int64Attribute{
			Description: "Field containing the IANA number.",
			Optional:    true,
		},
		"icmp_type": schema.Int64Attribute{
			Description: "Field containing the ICMP type.",
			Optional:    true,
		},
		"icmp_code": schema.Int64Attribute{
			Description: "Field containing the ICMP code.",
			Optional:    true,
		},
		"seed": schema.Int64Attribute{
			Description: communityIDSeedDescription,
			Optional:    true,
			Computed:    true,
			Validators:  []validator.Int64{int64validator.Between(0, 65535)},
		},
		"transport": schema.StringAttribute{
			Description: "Field containing the transport protocol. Used only when the `iana_number` field is not present.",
			Optional:    true,
		},
		attrTargetField: schema.StringAttribute{
			Description: "Output field for the community ID.",
			Optional:    true,
		},
		attrIgnoreMissing: schema.BoolAttribute{
			Description: descIgnoreMissingDocStop,
			Optional:    true,
			Computed:    true,
		},
	}

	maps.Copy(attrs, CommonProcessorSchemaAttributes())

	return NewProcessorDataSource(&processorCommunityIDModel{}, schema.Schema{
		Description: processorCommunityIDDataSourceDescription,
		Attributes:  attrs,
	})
}
