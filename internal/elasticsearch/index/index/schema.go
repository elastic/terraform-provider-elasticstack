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

package index

import (
	"context"
	"maps"

	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	indexparent "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/indexname"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/planmodifiers"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const dateMathIndexNameMessage = "must be a valid plain date math index name expression enclosed in angle brackets with at least one {…} section, e.g. <logs-{now/d}>"

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: resourceDescription,
		Blocks: map[string]schema.Block{
			"settings": schema.ListNestedBlock{
				Description:        deprecatedSettingsBlockDescription,
				DeprecationMessage: "Using settings makes it easier to misconfigure.  Use dedicated field for the each setting instead.",
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						attrSetting: schema.SetNestedBlock{
							Description: "Defines the setting for the index.",
							Validators: []validator.Set{
								setvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									attrName: schema.StringAttribute{
										Description: "The name of the setting to set and track.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "The value of the setting to set and track.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
		Attributes: getAttributes(),
	}
}

// getAttributes returns the index resource attribute map: the hand-declared
// static (creation-time-only) and operational attributes merged with the
// shared dynamic-setting attributes from indexparent.GetDynamicSettingAttributes().
func getAttributes() map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Internal identifier of the resource in the format <cluster_uuid>/<concrete_index_name>.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"concrete_name": schema.StringAttribute{
			Description: "The concrete Elasticsearch index name managed by this resource. " +
				"For static index names this equals `name`. " +
				"For date math index names this is the resolved concrete index name returned by Elasticsearch after creation.",
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Description: "Name of the index you wish to create.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.LengthBetween(1, 255),
				stringvalidator.NoneOf(".", ".."),
				stringvalidator.Any(
					stringvalidator.All(indexname.NameValidators()...),
					stringvalidator.RegexMatches(
						esclient.DateMathIndexNameRe,
						dateMathIndexNameMessage,
					),
				),
			},
		},
		"alias": schema.SetNestedAttribute{
			Description: "Aliases for the index.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.Set{
				setplanmodifier.UseStateForUnknown(),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Index alias name.",
						Required:    true,
					},
					attrFilter: schema.StringAttribute{
						Description: "Query used to limit documents the alias can access.",
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"index_routing": schema.StringAttribute{
						Description: "Value used to route indexing operations to a specific shard. If specified, this overwrites the `routing` value for indexing operations.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							planmodifiers.StringUseDefaultIfUnknown(""),
						},
					},
					"is_hidden": schema.BoolAttribute{
						Description: "If true, the alias is hidden.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							planmodifiers.BoolUseDefaultIfUnknown(false),
						},
					},
					"is_write_index": schema.BoolAttribute{
						Description: "If true, the index is the write index for the alias.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
							planmodifiers.BoolUseDefaultIfUnknown(false),
						},
					},
					"routing": schema.StringAttribute{
						Description: "Value used to route indexing and search operations to a specific shard.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							planmodifiers.StringUseDefaultIfUnknown(""),
						},
					},
					"search_routing": schema.StringAttribute{
						Description: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							planmodifiers.StringUseDefaultIfUnknown(""),
						},
					},
				},
			},
		},
		// Static settings that can only be set on creation
		indexparent.SettingNumberOfShards: schema.Int64Attribute{
			Description: "Number of shards for the index. This can be set only on creation.",
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		indexparent.SettingNumberOfRoutingShards: schema.Int64Attribute{
			Description: "Value used with number_of_shards to route documents to a primary shard. This can be set only on creation.",
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		indexparent.SettingCodec: schema.StringAttribute{
			Description: codecDescription,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("best_compression"),
			},
		},
		indexparent.SettingRoutingPartitionSize: schema.Int64Attribute{
			Description: "The number of shards a custom routing value can go to. This can be set only on creation.",
			Optional:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		indexparent.SettingLoadFixedBitsetFiltersEagerly: schema.BoolAttribute{
			Description: "Indicates whether cached filters are pre-loaded for nested queries. This can be set only on creation.",
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		"shard_check_on_startup": schema.StringAttribute{
			Description: shardCheckOnStartupDescription,
			Optional:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.OneOf("false", "true", "checksum"),
			},
		},
		"sort": schema.ListNestedAttribute{
			Description: "Sort configuration for documents within each shard segment. Replaces the deprecated sort_field and sort_order attributes.",
			Optional:    true,
			PlanModifiers: []planmodifier.List{
				sortMigrationPlanModifier{},
			},
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRoot("sort_field"), path.MatchRoot("sort_order")),
			},
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					attrField: schema.StringAttribute{
						Description: "The index field to sort by.",
						Required:    true,
					},
					attrOrder: schema.StringAttribute{
						Description: "The sort direction. Valid values: asc, desc.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(sortOrderAsc, sortOrderDesc),
						},
					},
					attrMissing: schema.StringAttribute{
						Description: "How to treat documents missing the sort field. Valid values: _last, _first.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(sortMissingLast, "_first"),
						},
					},
					attrMode: schema.StringAttribute{
						Description: "Which value to use when the sort field has multiple values. Valid values: min, max.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(sortModeMin, sortModeMax),
						},
					},
				},
			},
		},
		"sort_field": schema.SetAttribute{
			ElementType:        types.StringType,
			Description:        "Deprecated: The field to sort documents within each shard segment by.",
			Optional:           true,
			DeprecationMessage: "Use the 'sort' attribute instead. 'sort_field' will be removed in a future major release.",
			Validators: []validator.Set{
				setvalidator.ConflictsWith(path.MatchRoot("sort")),
			},
			PlanModifiers: []planmodifier.Set{
				legacySortFieldPlanModifier{},
			},
		},
		// sort_order can't be set type since it can have dup strings like ["asc", "asc"]
		"sort_order": schema.ListAttribute{
			ElementType:        types.StringType,
			Description:        "Deprecated: The direction to sort documents within each shard segment. Accepts `asc`, `desc`.",
			Optional:           true,
			DeprecationMessage: "Use the 'sort' attribute instead. 'sort_order' will be removed in a future major release.",
			Validators: []validator.List{
				listvalidator.ConflictsWith(path.MatchRoot("sort")),
			},
			PlanModifiers: []planmodifier.List{
				legacySortOrderPlanModifier{},
			},
		},
		"mapping_coerce": schema.BoolAttribute{
			Description: "Set index level coercion setting that is applied to all mapping types.",
			Optional:    true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		// To change analyzer setting, the index must be closed, updated, and then reopened but it can't be handled in terraform.
		// We raise error when they are tried to be updated instead of setting ForceNew not to have unexpected deletion.
		"analysis_analyzer": schema.StringAttribute{
			Description: "A JSON string describing the analyzers applied to the index.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
		},
		"analysis_tokenizer": schema.StringAttribute{
			Description: "A JSON string describing the tokenizers applied to the index.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
		},
		"analysis_char_filter": schema.StringAttribute{
			Description: "A JSON string describing the char_filters applied to the index.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
		},
		"analysis_filter": schema.StringAttribute{
			Description: "A JSON string describing the filters applied to the index.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
		},
		"analysis_normalizer": schema.StringAttribute{
			Description: "A JSON string describing the normalizers applied to the index.",
			Optional:    true,
			CustomType:  jsontypes.NormalizedType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
		},
		"mappings": schema.StringAttribute{
			Description: mappingsDescription,
			Optional:    true,
			Computed:    true,
			CustomType:  indexparent.MappingsType{},
			Validators: []validator.String{
				validators.StringIsJSONObject{},
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
				mappingsPlanModifier{},
			},
		},
		"settings_raw": schema.StringAttribute{
			Description: "All raw settings fetched from the cluster.",
			Computed:    true,
			CustomType:  jsontypes.NormalizedType{},
			// TODO: Plan modifier. Use state if no other settings have been modified
		},
		"deletion_protection": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: deletionProtectionDescription,
			PlanModifiers: []planmodifier.Bool{
				planmodifiers.BoolUseDefaultIfUnknown(true),
			},
		},
		"use_existing": schema.BoolAttribute{
			Description: useExistingDescription,
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"wait_for_active_shards": schema.StringAttribute{
			Description: waitForActiveShardsDescription,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				planmodifiers.StringUseDefaultIfUnknown("1"),
			},
		},
		"master_timeout": schema.StringAttribute{
			Description: masterTimeoutDescription,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				planmodifiers.StringUseDefaultIfUnknown("30s"),
			},
			CustomType: customtypes.DurationType{},
		},
		"timeout": schema.StringAttribute{
			Description: "Period to wait for a response. If no response is received before the timeout expires, the request fails and returns an error. Defaults to `30s`.",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				planmodifiers.StringUseDefaultIfUnknown("30s"),
			},
			CustomType: customtypes.DurationType{},
		},
	}

	maps.Copy(attributes, indexparent.GetDynamicSettingAttributes())

	return attributes
}

func aliasElementType(ctx context.Context) attr.Type {
	return getSchema(ctx).Attributes["alias"].GetType().(attr.TypeWithElementType).ElementType()
}

func settingsElementType(ctx context.Context) attr.Type {
	return getSchema(ctx).Blocks["settings"].Type().(attr.TypeWithElementType).ElementType()
}

func settingElementType(ctx context.Context) attr.Type {
	return getSchema(ctx).Blocks["settings"].GetNestedObject().GetBlocks()[attrSetting].Type().(attr.TypeWithElementType).ElementType()
}
