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

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/sslcore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	// keyAttributeName is the shared name of the ssl private-key attribute and
	// the headers header-name attribute; golangci requires a constant at three
	// occurrences.
	keyAttributeName = "key"

	sslAtLeastOneSubattrSummary = "Empty ssl configuration"
	sslAtLeastOneSubattrMessage = "At least one of certificate_authorities, certificate, key, or key_wo must be set."
)

var sslSubattrNames = []string{"certificate_authorities", "certificate", keyAttributeName, "key_wo"}

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
					sslcore.AtLeastOneSubattrValidator(sslSubattrNames, sslAtLeastOneSubattrSummary, sslAtLeastOneSubattrMessage),
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
					keyAttributeName: schema.StringAttribute{
						Description: "Client SSL certificate key for the download source host. Sent to Fleet as a managed secret.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"key_wo": schema.StringAttribute{
						Description: "Write-only client SSL certificate key for the download source host. Never stored in Terraform state; " +
							"changes are detected from a hash kept in private state. Conflicts with `key`.",
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(keyAttributeName)),
						},
					},
				},
			},
			"auth": schema.SingleNestedAttribute{
				Description: "Authentication used by Elastic Agents when downloading binaries from this source.",
				Optional:    true,
				Validators: []validator.Object{
					authCredentialsValidator{},
				},
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "Username for HTTP Basic authentication. Required when password or password_wo is set.",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password for HTTP Basic authentication. Required when username is set. Conflicts with password_wo and the api_key attributes.",
						Optional:    true,
						Sensitive:   true,
					},
					"password_wo": schema.StringAttribute{
						Description: "Write-only password for HTTP Basic authentication. Never stored in Terraform state; changes are detected " +
							"from a hash kept in private state. Required when username is set. Conflicts with password and the api_key attributes.",
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
					},
					"api_key": schema.StringAttribute{
						Description: "API key used to authenticate download requests. Conflicts with api_key_wo and the password attributes.",
						Optional:    true,
						Sensitive:   true,
					},
					"api_key_wo": schema.StringAttribute{
						Description: "Write-only API key used to authenticate download requests. Never stored in Terraform state; changes are " +
							"detected from a hash kept in private state. Conflicts with api_key and the password attributes.",
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
					},
					"headers": schema.ListNestedAttribute{
						Description: "Custom HTTP headers sent with every download request.",
						Optional:    true,
						Validators: []validator.List{
							listvalidator.SizeAtMost(100),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								keyAttributeName: schema.StringAttribute{
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
