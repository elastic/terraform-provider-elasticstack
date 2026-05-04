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
	esindex "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema() dschema.Schema {
	return dschema.Schema{
		MarkdownDescription: mdDescIndexTemplateDataSource,
		Blocks: map[string]dschema.Block{
			"data_stream": dataSourceDataStreamBlock(),
			"template":    dataSourceTemplateBlock(),
		},
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				MarkdownDescription: descID,
				Computed:            true,
			},
			"name": dschema.StringAttribute{
				MarkdownDescription: descNameDataSrc,
				Required:            true,
			},
			"composed_of": dschema.ListAttribute{
				MarkdownDescription: descComposedOf,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ignore_missing_component_templates": dschema.ListAttribute{
				MarkdownDescription: descIgnoreMissingComponentTemplates,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"index_patterns": dschema.SetAttribute{
				MarkdownDescription: descIndexPatterns,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"metadata": dschema.StringAttribute{
				MarkdownDescription: descMetadata,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"priority": dschema.Int64Attribute{
				MarkdownDescription: descPriority,
				Computed:            true,
			},
			"version": dschema.Int64Attribute{
				MarkdownDescription: descVersion,
				Computed:            true,
			},
		},
	}
}

func dataSourceDataStreamBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descDataStreamBlock,
		Attributes: map[string]dschema.Attribute{
			"hidden": dschema.BoolAttribute{
				MarkdownDescription: descDataStreamHidden,
				Computed:            true,
			},
			"allow_custom_routing": dschema.BoolAttribute{
				MarkdownDescription: descDataStreamAllowCustomRouting,
				Computed:            true,
			},
		},
	}
}

func dataSourceTemplateBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descTemplateBlock,
		Attributes: map[string]dschema.Attribute{
			"mappings": dschema.StringAttribute{
				MarkdownDescription: descTemplateMappings,
				Computed:            true,
				CustomType:          esindex.MappingsType{},
			},
			"settings": dschema.StringAttribute{
				MarkdownDescription: descTemplateSettings,
				Computed:            true,
				CustomType:          customtypes.IndexSettingsType{},
			},
		},
		Blocks: map[string]dschema.Block{
			"alias":               dataSourceTemplateAliasBlock(),
			"lifecycle":           dataSourceTemplateLifecycleBlock(),
			"data_stream_options": dataSourceTemplateDataStreamOptionsBlock(),
		},
	}
}

func dataSourceTemplateAliasBlock() dschema.SetNestedBlock {
	return dschema.SetNestedBlock{
		MarkdownDescription: descAliasBlock,
		NestedObject: dschema.NestedBlockObject{
			CustomType: NewAliasObjectType(),
			Attributes: map[string]dschema.Attribute{
				"name": dschema.StringAttribute{
					MarkdownDescription: descAliasName,
					Computed:            true,
				},
				"filter": dschema.StringAttribute{
					MarkdownDescription: descAliasFilter,
					Computed:            true,
					CustomType:          jsontypes.NormalizedType{},
				},
				"index_routing": dschema.StringAttribute{
					MarkdownDescription: descAliasIndexRouting,
					Computed:            true,
				},
				"is_hidden": dschema.BoolAttribute{
					MarkdownDescription: descAliasIsHidden,
					Computed:            true,
				},
				"is_write_index": dschema.BoolAttribute{
					MarkdownDescription: descAliasIsWriteIndex,
					Computed:            true,
				},
				"routing": dschema.StringAttribute{
					MarkdownDescription: descAliasRouting,
					Computed:            true,
				},
				"search_routing": dschema.StringAttribute{
					MarkdownDescription: descAliasSearchRouting,
					Computed:            true,
				},
			},
		},
	}
}

func dataSourceTemplateLifecycleBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descLifecycleBlock,
		Attributes: map[string]dschema.Attribute{
			"data_retention": dschema.StringAttribute{
				MarkdownDescription: descLifecycleDataRetention,
				Computed:            true,
			},
		},
	}
}

func dataSourceTemplateDataStreamOptionsBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descDataStreamOptionsBlockDataSource,
		Blocks: map[string]dschema.Block{
			"failure_store": dataSourceTemplateFailureStoreBlock(),
		},
	}
}

func dataSourceTemplateFailureStoreBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descFailureStoreBlock,
		Attributes: map[string]dschema.Attribute{
			"enabled": dschema.BoolAttribute{
				MarkdownDescription: descFailureStoreEnabled,
				Computed:            true,
			},
		},
		Blocks: map[string]dschema.Block{
			"lifecycle": dataSourceTemplateFailureStoreLifecycleBlock(),
		},
	}
}

func dataSourceTemplateFailureStoreLifecycleBlock() dschema.SingleNestedBlock {
	return dschema.SingleNestedBlock{
		MarkdownDescription: descFailureStoreLifecycleBlock,
		Attributes: map[string]dschema.Attribute{
			"data_retention": dschema.StringAttribute{
				MarkdownDescription: descFailureStoreDataRetention,
				Computed:            true,
			},
		},
	}
}
