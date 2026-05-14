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

package componenttemplate

import (
	"context"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// aliasAttrTypes returns the attribute types for a single alias nested block element.
func aliasAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"filter":         jsontypes.NormalizedType{},
		"index_routing":  types.StringType,
		"is_hidden":      types.BoolType,
		"is_write_index": types.BoolType,
		"routing":        types.StringType,
		"search_routing": types.StringType,
	}
}

// templateAttrTypes returns the attribute types for the template block object.
func templateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"alias":    types.SetType{ElemType: types.ObjectType{AttrTypes: aliasAttrTypes()}},
		"mappings": esindex.MappingsType{},
		"settings": customtypes.IndexSettingsType{},
	}
}

// getSchema returns the Plugin Framework schema for elasticstack_elasticsearch_component_template.
// The elasticsearch_connection block is NOT included here; the envelope injects it.
const schemaVersion int64 = 1

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Version: schemaVersion,
		MarkdownDescription: "Creates or updates a component template. Component templates are building blocks for constructing index templates " +
			"that specify index mappings, settings, and aliases. See the " +
			"[component template documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-component-template.html) " +
			"for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the component template to create.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional user metadata about the component template.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "Version number used to manage component templates externally.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"template": schema.SingleNestedBlock{
				MarkdownDescription: "Template to be applied. It may optionally include an aliases, mappings, or settings configuration.",
				Attributes: map[string]schema.Attribute{
					"mappings": schema.StringAttribute{
						MarkdownDescription: "Mapping for fields in the index. Should be specified as a JSON object of field mappings. " +
							"See the [explicit mapping documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/explicit-mapping.html) " +
							"for more details.",
						Optional:   true,
						CustomType: esindex.MappingsType{},
						Validators: []validator.String{
							esindex.StringIsJSONObject{},
						},
					},
					"settings": schema.StringAttribute{
						MarkdownDescription: "Configuration options for the index. See the " +
							"[index modules settings documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings) " +
							"for more details.",
						Optional:   true,
						CustomType: customtypes.IndexSettingsType{},
					},
				},
				Blocks: map[string]schema.Block{
					"alias": schema.SetNestedBlock{
						MarkdownDescription: "Alias to add.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The alias name. Index alias names support date math. See the " +
										"[date math index names documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/date-math-index-names.html) " +
										"for more details.",
									Required: true,
								},
								"filter": schema.StringAttribute{
									MarkdownDescription: "Query used to limit documents the alias can access.",
									Optional:            true,
									CustomType:          jsontypes.NormalizedType{},
									Validators: []validator.String{
										esindex.StringIsJSONObject{},
									},
								},
								"index_routing": schema.StringAttribute{
									MarkdownDescription: "Value used to route indexing operations to a specific shard. If specified, this overwrites the routing value for indexing operations.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
								},
								"is_hidden": schema.BoolAttribute{
									MarkdownDescription: "If true, the alias is hidden.",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"is_write_index": schema.BoolAttribute{
									MarkdownDescription: "If true, the index is the write index for the alias.",
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
								},
								"routing": schema.StringAttribute{
									MarkdownDescription: "Value used to route indexing and search operations to a specific shard.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
								},
								"search_routing": schema.StringAttribute{
									MarkdownDescription: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
								},
							},
						},
					},
				},
			},
		},
	}
}
