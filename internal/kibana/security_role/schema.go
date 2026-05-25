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

package security_role

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed resource-description.md
var resourceDescription string

//go:embed descriptions/remote_indices_permissions.md
var remoteIndicesPermissionsDescription string

func fieldSecurityResourceAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrGrant: schema.SetAttribute{
			Description: descFieldSecurityGrant,
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		attrExcept: schema.SetAttribute{
			Description: descFieldSecurityExcept,
			Optional:    true,
			ElementType: types.StringType,
		},
	}
}

func fieldSecurityResourceBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: descFieldSecurityBlock,
		Attributes:  fieldSecurityResourceAttrs(),
	}
}

func commonIndexBlockAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		attrNames: schema.SetAttribute{
			Description: descIndexNames,
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		attrPrivileges: schema.SetAttribute{
			Description: descIndexPrivileges,
			Required:    true,
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		attrQuery: schema.StringAttribute{
			Description: descIndexQuery,
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
		},
	}
}

func indicesResourceBlock() schema.Block {
	return schema.SetNestedBlock{
		Description: "A list of indices permissions entries.",
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				attrFieldSecurity: fieldSecurityResourceBlock(),
			},
			Attributes: commonIndexBlockAttrs(),
		},
	}
}

func remoteIndicesResourceBlock() schema.Block {
	attrs := commonIndexBlockAttrs()
	attrs[attrClusters] = schema.SetAttribute{
		Description: "A list of cluster aliases to which the permissions in this entry apply.",
		Required:    true,
		ElementType: types.StringType,
	}
	return schema.SetNestedBlock{
		Description: remoteIndicesPermissionsDescription,
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				attrFieldSecurity: fieldSecurityResourceBlock(),
			},
			Attributes: attrs,
		},
	}
}

func getResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: resourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch": schema.SingleNestedBlock{
				Description: "Elasticsearch cluster and index privileges.",
				Attributes: map[string]schema.Attribute{
					attrCluster: schema.SetAttribute{
						Description: "List of the cluster privileges.",
						Optional:    true,
						ElementType: types.StringType,
					},
					attrRunAs: schema.SetAttribute{
						Description: "A list of usernames the owners of this role can impersonate.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					attrIndices:       indicesResourceBlock(),
					attrRemoteIndices: remoteIndicesResourceBlock(),
				},
			},
			"kibana": schema.SetNestedBlock{
				Description: "The list of objects that specify the Kibana privileges for the role.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						attrSpaces: schema.SetAttribute{
							Description: "The spaces to apply the privileges to. To grant access to all spaces, set to [\"*\"], or omit the value.",
							Required:    true,
							ElementType: types.StringType,
						},
						attrBase: schema.SetAttribute{
							Description: "A base privilege. When specified, the base must be [\"all\"] or [\"read\"]. When the base privileges are specified, you are unable to use the \"feature\" section.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtMost(1),
								setvalidator.ValueStringsAre(
									stringvalidator.RegexMatches(
										regexp.MustCompile(`(?i)^(all|read)$`),
										"must be 'all' or 'read'",
									),
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						attrFeature: schema.SetNestedBlock{
							Description: "List of privileges for specific features. When the feature privileges are specified, you are unable to use the \"base\" section.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									attrName: schema.StringAttribute{
										Description: "Feature name.",
										Required:    true,
									},
									attrPrivileges: schema.SetAttribute{
										Description: "Feature privileges.",
										Required:    true,
										ElementType: types.StringType,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			attrName: schema.StringAttribute{
				Description: "The name for the role.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "Internal identifier (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description for the role",
				Optional:    true,
			},
			"metadata": schema.StringAttribute{
				Description: "Optional meta-data.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
