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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const spaceIDsDescription = "The Kibana space IDs where this download source is available. When set, the download source will be created " +
	"and managed within the specified space. Note: The order of space IDs does not matter as this is a set."

const (
	attrDefault = "default"
)

func getSchema(_ context.Context) schema.Schema {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")

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
			"auth": schema.SingleNestedAttribute{
				Description: "Authentication used by Elastic Agents when downloading binaries from this source.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Username for HTTP Basic authentication. Required when password is set.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.AlsoRequires(passwordPath),
						},
					},
					"password": schema.StringAttribute{
						Description: "Password for HTTP Basic authentication. Required when username is set.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.AlsoRequires(usernamePath),
						},
					},
					"api_key": schema.StringAttribute{
						Description: "API key used to authenticate download requests. Conflicts with password.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(passwordPath),
						},
					},
					"headers": schema.ListNestedAttribute{
						Description: "Custom HTTP headers sent with every download request.",
						Optional:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Description: "HTTP header name.",
									Required:    true,
								},
								"value": schema.StringAttribute{
									Description: "HTTP header value.",
									Required:    true,
								},
							},
						},
					},
				},
			},
			"space_ids": kbschema.SpaceIDsAttribute(spaceIDsDescription),
		},
	}
}
