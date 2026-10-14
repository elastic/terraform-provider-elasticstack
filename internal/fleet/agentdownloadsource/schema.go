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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const spaceIDsDescription = "The Kibana space IDs where this download source is available. When set, the download source will be created " +
	"and managed within the specified space. Note: The order of space IDs does not matter as this is a set."

const (
	attrDefault = "default"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Creates a new Fleet Agent Binary Download Source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_id": schema.StringAttribute{
				Description: "Unique identifier of the Fleet agent download source.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Fleet agent download source.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "The base URL from which Elastic Agents will download binaries.",
				Required:    true,
			},
			attrDefault: schema.BoolAttribute{
				Description: "Set this download source as the default for agents.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"proxy_id": schema.StringAttribute{
				Description: "The ID of the proxy to use for this download source.",
				Optional:    true,
			},
			"ssl": schema.SingleNestedAttribute{
				Description: "SSL/TLS configuration for the download source host.",
				Optional:    true,
				Validators: []validator.Object{
					sslAtLeastOneSubattrValidator{},
				},
				Attributes: map[string]schema.Attribute{
					"certificate_authorities": schema.ListAttribute{
						Description: "Server SSL certificate authorities trusted for the download source host.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
							listvalidator.SizeAtMost(10),
							listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
						},
					},
					"certificate": schema.StringAttribute{
						Description: "Client SSL certificate for the download source host.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"key": schema.StringAttribute{
						Description: "Client SSL certificate key for the download source host.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
			},
			"space_ids": kbschema.SpaceIDsAttribute(spaceIDsDescription),
		},
	}
}
