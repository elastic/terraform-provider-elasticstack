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

package template

import (
	"context"

	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema()
}

func resourceSchema() schema.Schema {
	return schema.Schema{
		Version:             schemaVersion,
		MarkdownDescription: mdDescIndexTemplateResource,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(),
			"data_stream":              dataStreamBlock(),
			"template":                 templateBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: descID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: descName,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"composed_of": schema.ListAttribute{
				MarkdownDescription: descComposedOf,
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"ignore_missing_component_templates": schema.ListAttribute{
				MarkdownDescription: descIgnoreMissingComponentTemplates,
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"index_patterns": schema.SetAttribute{
				MarkdownDescription: descIndexPatterns,
				Required:            true,
				ElementType:         types.StringType,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: descMetadata,
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: descPriority,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: descVersion,
				Optional:            true,
			},
		},
	}
}

func dataStreamBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descDataStreamBlock,
		Attributes: map[string]schema.Attribute{
			"hidden": schema.BoolAttribute{
				MarkdownDescription: descDataStreamHidden,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"allow_custom_routing": schema.BoolAttribute{
				MarkdownDescription: descDataStreamAllowCustomRouting,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func templateBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descTemplateBlock,
		Attributes: map[string]schema.Attribute{
			"mappings": schema.StringAttribute{
				MarkdownDescription: descTemplateMappings,
				Optional:            true,
				CustomType:          esindex.MappingsType{},
				Validators: []validator.String{
					esindex.StringIsJSONObject{},
				},
			},
			"settings": schema.StringAttribute{
				MarkdownDescription: descTemplateSettings,
				Optional:            true,
				CustomType:          customtypes.IndexSettingsType{},
			},
		},
		Blocks: map[string]schema.Block{
			"alias":               templateAliasBlock(),
			"lifecycle":           templateLifecycleBlock(),
			"data_stream_options": templateDataStreamOptionsBlock(),
		},
	}
}

func templateAliasBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: descAliasBlock,
		NestedObject: schema.NestedBlockObject{
			CustomType: NewAliasObjectType(),
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: descAliasName,
					Required:            true,
				},
				"filter": schema.StringAttribute{
					MarkdownDescription: descAliasFilter,
					Optional:            true,
					CustomType:          jsontypes.NormalizedType{},
				},
				"index_routing": schema.StringAttribute{
					MarkdownDescription: descAliasIndexRouting,
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
				"is_hidden": schema.BoolAttribute{
					MarkdownDescription: descAliasIsHidden,
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"is_write_index": schema.BoolAttribute{
					MarkdownDescription: descAliasIsWriteIndex,
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				"routing": schema.StringAttribute{
					MarkdownDescription: descAliasRouting,
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
				"search_routing": schema.StringAttribute{
					MarkdownDescription: descAliasSearchRouting,
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
			},
		},
	}
}

func templateLifecycleBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descLifecycleBlock,
		Attributes: map[string]schema.Attribute{
			"data_retention": schema.StringAttribute{
				MarkdownDescription: descLifecycleDataRetention,
				Optional:            true,
			},
		},
	}
}

func templateDataStreamOptionsBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descDataStreamOptionsBlock,
		Blocks: map[string]schema.Block{
			"failure_store": templateFailureStoreBlock(),
		},
	}
}

func templateFailureStoreBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descFailureStoreBlock,
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: descFailureStoreEnabled,
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"lifecycle": templateFailureStoreLifecycleBlock(),
		},
	}
}

func templateFailureStoreLifecycleBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descFailureStoreLifecycleBlock,
		Attributes: map[string]schema.Attribute{
			"data_retention": schema.StringAttribute{
				MarkdownDescription: descFailureStoreDataRetention,
				Optional:            true,
			},
		},
	}
}
