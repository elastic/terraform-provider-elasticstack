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

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/sslcore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	sslCertificateAttr            = "certificate"
	sslKeyAttr                    = "key"
	sslCertificateAuthoritiesAttr = "certificate_authorities"
	sslClientAuthAttr             = "client_auth"

	sslAtLeastOneSubattrSummary = "Empty ssl configuration"
	sslAtLeastOneSubattrMessage = "At least one of certificate, key, certificate_authorities, or client_auth must be set."
)

var sslSubattrNames = []string{sslCertificateAttr, sslKeyAttr, sslCertificateAuthoritiesAttr, sslClientAuthAttr}

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Creates a new Fleet Server Host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"host_id": schema.StringAttribute{
				Description: `Unique identifier of the Fleet server host. When omitted, Fleet auto-generates an ID. ` +
					`When set, the value must be 1-255 characters and must not contain path separators ("/"), ` +
					`traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype"). ` +
					`Invalid explicit values fail at plan time.`,
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fleet.IDValidator("host_id"),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Fleet server host.",
				Required:    true,
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"default": schema.BoolAttribute{
				Description: "Set as default.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"proxy_id": schema.StringAttribute{
				Description: "The ID of the Fleet proxy to use for this Fleet server host. Must not be empty; omit the attribute to use no proxy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ssl": schema.SingleNestedAttribute{
				Description: sslDescription,
				Optional:    true,
				Validators: []validator.Object{
					sslcore.AtLeastOneSubattrValidator(sslSubattrNames, sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage),
				},
				Attributes: map[string]schema.Attribute{
					sslCertificateAttr: schema.StringAttribute{
						Description: "The TLS certificate Fleet Server presents to Elastic Agents.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					sslKeyAttr: schema.StringAttribute{
						Description: "The private key for the Fleet Server TLS certificate.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					sslCertificateAuthoritiesAttr: schema.ListAttribute{
						Description: "Certificate authorities used to validate Elastic Agent client certificates. Accepts 1-10 entries.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.SizeAtMost(10),
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					sslClientAuthAttr: schema.StringAttribute{
						Description: "Whether Fleet Server requests client certificates from Elastic Agents. Valid values are `none`, `optional`, and `required`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "optional", "required"),
						},
					},
				},
			},
			"space_ids": kbschema.SpaceIDsAttribute(spaceIDsDescription),
		},
	}
}

var sslAttrTypes = map[string]attr.Type{
	sslCertificateAttr:            types.StringType,
	sslKeyAttr:                    types.StringType,
	sslCertificateAuthoritiesAttr: types.ListType{ElemType: types.StringType},
	sslClientAuthAttr:             types.StringType,
}
