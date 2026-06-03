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

package entities

import (
	"context"

	entity "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entity"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Queries the Kibana Security Entity Store list/search endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Stable identifier computed as <space_id>/entity_store_entities.",
				Computed:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the Kibana space. If omitted, the default space is used.",
				Optional:    true,
				Computed:    true,
			},
			"entity_id": schema.StringAttribute{
				Description: "When set, the provider generates an implicit KQL filter for this entity id. Conflicts with filter and filter_query.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("filter"), path.MatchRoot("filter_query")),
				},
			},
			"filter": schema.StringAttribute{
				Description: "A Kibana Query Language (KQL) filter for the search-after mode.",
				Optional:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Number of entities to return in search-after mode.",
				Optional:    true,
			},
			"search_after": schema.StringAttribute{
				Description: "JSON-encoded search_after cursor from a previous response.",
				Optional:    true,
			},
			"source": schema.ListAttribute{
				Description: "Fields to include in response _source.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"fields": schema.ListAttribute{
				Description: "Fields to include in response fields.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"sort_field": schema.StringAttribute{
				Description: "Field to sort results by in page mode.",
				Optional:    true,
			},
			"sort_order": schema.StringAttribute{
				Description: "Sort order in page mode.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("asc", "desc"),
				},
			},
			"page": schema.Int64Attribute{
				Description: "Page number to return (1-indexed) in page mode.",
				Optional:    true,
			},
			"per_page": schema.Int64Attribute{
				Description: "Number of entities per page in page mode.",
				Optional:    true,
			},
			"filter_query": schema.StringAttribute{
				Description: "An Elasticsearch query string to filter entities in page mode.",
				Optional:    true,
			},
			"entity_types": schema.SetAttribute{
				Description: "Entity types to include in the results. Valid values are user, host, service, generic.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf("user", "host", "service", "generic")),
				},
			},
			"results_json": schema.StringAttribute{
				Description: "Normalized JSON (sorted keys) of the full API response body.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"items": schema.ListAttribute{
				Description: "List of entity records with typed attributes matching the resource schema.",
				Computed:    true,
				ElementType: entity.ItemObjectType(),
			},
		},
	}
}

// expandStringList converts a types.List to a []string.
func expandStringList(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	result := make([]string, 0, len(l.Elements()))
	for _, v := range l.Elements() {
		if s, ok := v.(types.String); ok {
			result = append(result, s.ValueString())
		}
	}
	return result
}
