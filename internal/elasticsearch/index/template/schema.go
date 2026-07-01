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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/aliasutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Version:             schemaVersion,
		MarkdownDescription: mdDescIndexTemplateResource,
		Blocks: map[string]schema.Block{
			attrDataStream: dataStreamBlock(),
			attrTemplate:   templateBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: descID,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrName: schema.StringAttribute{
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
			attrIndexPatterns: schema.SetAttribute{
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
			"allow_auto_create": schema.BoolAttribute{
				MarkdownDescription: descAllowAutoCreate,
				Optional:            true,
			},
		},
	}
}

func dataStreamBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descDataStreamBlock,
		Attributes: map[string]schema.Attribute{
			attrHidden: schema.BoolAttribute{
				MarkdownDescription: descDataStreamHidden,
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			attrAllowCustomRouting: schema.BoolAttribute{
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
			attrMappings: schema.StringAttribute{
				MarkdownDescription: descTemplateMappings,
				Optional:            true,
				CustomType:          esindex.MappingsType{},
				Validators: []validator.String{
					validators.StringIsJSONObject{},
				},
			},
			attrSettings: schema.StringAttribute{
				MarkdownDescription: descTemplateSettings,
				Optional:            true,
				CustomType:          customtypes.IndexSettingsType{},
			},
		},
		Blocks: map[string]schema.Block{
			attrAlias:             aliasutil.AliasSetNestedBlock(),
			attrLifecycle:         templateLifecycleBlock(),
			attrDataStreamOptions: datastreamoptions.Block(),
		},
	}
}

func templateLifecycleBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: descLifecycleBlock,
		Attributes: map[string]schema.Attribute{
			attrDataRetention: schema.StringAttribute{
				MarkdownDescription: descLifecycleDataRetention,
				Optional:            true,
			},
		},
	}
}
