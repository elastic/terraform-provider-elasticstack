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

package security_entity_store

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages the Elastic Security Entity Store lifecycle within a Kibana space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Computed resource identifier in the format <space_id>/entity_store.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the Kibana space. If omitted, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultSpaceID),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_types": schema.SetAttribute{
				Description: "Entity types to install and manage. Valid values are user, host, service, and generic.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf("user", "host", "service", "generic")),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_entity_type_shrink": schema.BoolAttribute{
				Description: "Terraform-only guard that permits removing installed entity types when true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"started": schema.BoolAttribute{
				Description: "Whether any managed entity engine should be running after reconciliation.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"history_snapshot": schema.SingleNestedAttribute{
				Description: "Install-only history snapshot settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					frequencyAttr: schema.StringAttribute{
						Description: "History snapshot frequency used during installation.",
						Optional:    true,
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"log_extraction": schema.SingleNestedAttribute{
				Description: "Optional log extraction settings for the entity store.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"additional_index_patterns": schema.ListAttribute{
						Optional:      true,
						Computed:      true,
						ElementType:   types.StringType,
						PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					},
					"excluded_index_patterns": schema.ListAttribute{
						Optional:      true,
						Computed:      true,
						ElementType:   types.StringType,
						PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
					},
					"delay": schema.StringAttribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"docs_limit": schema.Int64Attribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"field_history_length": schema.Int64Attribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"frequency": schema.StringAttribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"lookback_period": schema.StringAttribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"max_logs_per_page": schema.Int64Attribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"max_logs_per_window": schema.Int64Attribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
					},
					"max_logs_per_window_cap_behavior": schema.StringAttribute{
						Optional: true,
						Computed: true,
						Validators: []validator.String{
							stringvalidator.OneOf("drop", "defer"),
						},
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"max_time_window_size": schema.StringAttribute{
						Optional:      true,
						Computed:      true,
						PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"status_json": schema.StringAttribute{
				Description: "Normalized JSON representation of the most recent entity store status response.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
