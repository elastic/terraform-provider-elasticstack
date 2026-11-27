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

package serverhostds

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	hostIDAttr  = "host_id"
	nameAttr    = "name"
	defaultAttr = "default"

	sslCertificateAttr            = "certificate"
	sslCertificateAuthoritiesAttr = "certificate_authorities"
	sslClientAuthAttr             = "client_auth"

	selectorDescription = "Set exactly one of `host_id`, `name` or `default`; the others are populated from the API response."
)

func getDataSourceSchema(_ context.Context) dsschema.Schema {
	return dsschema.Schema{
		Description: "Returns a Fleet server host looked up by `host_id`, by `name`, or as the default Fleet server host (`default = true`).",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The ID of this Fleet server host.",
				Computed:    true,
			},
			hostIDAttr: dsschema.StringAttribute{
				Description: "The identifier of the Fleet server host. " + selectorDescription,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(
						path.MatchRoot(hostIDAttr),
						path.MatchRoot(nameAttr),
						path.MatchRoot(defaultAttr),
					),
				},
			},
			nameAttr: dsschema.StringAttribute{
				Description: "The name of the Fleet server host. " + selectorDescription,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			defaultAttr: dsschema.BoolAttribute{
				Description: "Set to `true` to look up the default Fleet server host; `false` is not a valid selector. " +
					"When `host_id` or `name` is the selector, this reports whether the matched host is the default one. " +
					selectorDescription,
				Optional: true,
				Computed: true,
				Validators: []validator.Bool{
					boolvalidator.Equals(true),
				},
			},
			"space_id": spaceIDAttribute(),
			"hosts": dsschema.ListAttribute{
				Description: "The URLs of the Fleet server host.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"proxy_id": dsschema.StringAttribute{
				Description: "The ID of the Fleet proxy used by this Fleet server host.",
				Computed:    true,
			},
			"ssl": dsschema.SingleNestedAttribute{
				Description: "TLS configuration Fleet Server presents to Elastic Agents. The private key is not exposed.",
				Computed:    true,
				Attributes: map[string]dsschema.Attribute{
					sslCertificateAttr: dsschema.StringAttribute{
						Description: "The TLS certificate Fleet Server presents to Elastic Agents.",
						Computed:    true,
					},
					sslCertificateAuthoritiesAttr: dsschema.ListAttribute{
						Description: "Certificate authorities used to validate Elastic Agent client certificates.",
						ElementType: types.StringType,
						Computed:    true,
					},
					sslClientAuthAttr: dsschema.StringAttribute{
						Description: "Whether Fleet Server requests client certificates from Elastic Agents: `none`, `optional` or `required`.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func spaceIDAttribute() dsschema.StringAttribute {
	attribute := kbschema.DataSourceSpaceIDAttribute()
	attribute.MarkdownDescription = "The Kibana space the Fleet API request is sent through. " +
		"Fleet server hosts are shared by all spaces, so `space_id` does not filter the lookup. " +
		"If `space_id` is not provided, the default space is used."
	return attribute
}

var sslAttrTypes = map[string]attr.Type{
	sslCertificateAttr:            types.StringType,
	sslCertificateAuthoritiesAttr: types.ListType{ElemType: types.StringType},
	sslClientAuthAttr:             types.StringType,
}
