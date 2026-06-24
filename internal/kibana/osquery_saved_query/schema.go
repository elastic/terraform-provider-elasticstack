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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var osqueryPlatformValues = []string{"linux", "darwin", "windows"}

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a user-defined Osquery saved query in Kibana. " +
			"Prebuilt queries shipped with the osquery_manager integration cannot be managed by this resource; " +
			"use the `elasticstack_kibana_osquery_saved_query` data source to read them instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier in the form `<space_id>/<saved_query_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"saved_query_id": schema.StringAttribute{
				MarkdownDescription: "Stable identifier for the saved query. Used as the API lookup key and forces replacement when changed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "Kibana space identifier. When omitted, the default space is used.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(clients.DefaultSpaceID),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"query": schema.StringAttribute{
				MarkdownDescription: "Osquery SQL query text.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Human-readable description of the saved query.",
				Optional:            true,
			},
			"platform": schema.SetAttribute{
				MarkdownDescription: "Target platforms for the query. Allowed values: `linux`, `darwin`, `windows`.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(osqueryPlatformValues...)),
				},
			},
			"interval": schema.Int64Attribute{
				MarkdownDescription: "Query execution interval in seconds.",
				Optional:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Saved query version string.",
				Optional:            true,
			},
			"snapshot": schema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is a snapshot. Returned by the API and may be set explicitly in configuration. " +
					"When omitted or unknown at plan time, the prior state value is preserved (`UseStateForUnknown`).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"removed": schema.BoolAttribute{
				MarkdownDescription: "Whether the saved query is marked removed. Returned by the API and may be set explicitly in configuration. " +
					"When omitted or unknown at plan time, the prior state value is preserved (`UseStateForUnknown`).",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ecs_mapping": ecsMappingSchema(),
		},
	}
}

func ecsMappingSchema() schema.MapNestedAttribute {
	return schema.MapNestedAttribute{
		MarkdownDescription: "Maps query result columns to ECS field paths. Each map value must set exactly one of `field`, `value`, or `values`.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Validators: []validator.Object{
				ecsMappingExactlyOneOfValidator(),
			},
			Attributes: map[string]schema.Attribute{
				attrEcsMappingField: schema.StringAttribute{
					MarkdownDescription: "Query result column name to map from.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				attrEcsMappingValue: schema.StringAttribute{
					MarkdownDescription: "Static scalar ECS mapping value.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				attrEcsMappingValues: schema.SetAttribute{
					MarkdownDescription: "Static array ECS mapping values.",
					Optional:            true,
					ElementType:         types.StringType,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
				},
			},
		},
	}
}

// ecsMappingExactlyOneOfValidator enforces exactly one of field/value/values per ecs_mapping
// element. Uses ExactlyOneOfNestedAttrsValidator (primary path per design Decision 6); a custom
// inline ValidateObject fallback is documented in tasks.md if map nested validation fails in CI.
func ecsMappingExactlyOneOfValidator() validator.Object {
	return validators.ExactlyOneOfNestedAttrsValidator(validators.ExactlyOneOfNestedAttrsOpts{
		AttrNames:     []string{attrEcsMappingField, attrEcsMappingValue, attrEcsMappingValues},
		Summary:       "Invalid ecs_mapping element",
		MissingDetail: "Exactly one of `field`, `value`, or `values` must be set per `ecs_mapping` element.",
		TooManyDetail: "Exactly one of `field`, `value`, or `values` must be set per `ecs_mapping` element, not more than one.",
		Description:   "Ensures exactly one of `field`, `value`, or `values` is set on each `ecs_mapping` map value.",
	})
}
