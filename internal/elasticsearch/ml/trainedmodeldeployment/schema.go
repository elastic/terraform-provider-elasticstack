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

package trainedmodeldeployment

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func GetSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource in the format `<cluster_uuid>/<deployment_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the trained model to deploy.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment_id": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the deployment of the model. Defaults to the value of `model_id`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"number_of_allocations": schema.Int64Attribute{
				MarkdownDescription: "The number of model allocations on each node where the model is deployed. Cannot be set when `adaptive_allocations` is configured.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("adaptive_allocations")),
				},
			},
			"threads_per_allocation": schema.Int64Attribute{
				MarkdownDescription: "Sets the number of threads used by each model allocation during inference.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "The deployment priority. Valid values are `low` and `normal`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("low", "normal"),
				},
			},
			"queue_capacity": schema.Int64Attribute{
				MarkdownDescription: "Specifies the number of inference requests that are allowed in the queue.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"wait_for": schema.StringAttribute{
				MarkdownDescription: "Specifies the allocation status to wait for before returning. Valid values are `starting`, `started`, and `fully_allocated`. Defaults to `fully_allocated`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("fully_allocated"),
				Validators: []validator.String{
					stringvalidator.OneOf("starting", "started", "fully_allocated"),
				},
			},
			"api_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the amount of time to wait for the model to deploy. This is the server-side start timeout.",
				Optional:            true,
				CustomType:          customtypes.DurationType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_stop": schema.BoolAttribute{
				MarkdownDescription: "When `true`, passes `force=true` to the Stop Deployment API on destroy.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The overall state of the deployment.",
				Computed:            true,
			},
			"allocation_status": schema.StringAttribute{
				MarkdownDescription: "The detailed allocation state of the deployment.",
				Computed:            true,
			},
			"stats_json": schema.StringAttribute{
				MarkdownDescription: "The raw JSON of the trained model stats for this deployment.",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
			"adaptive_allocations": schema.SingleNestedAttribute{
				MarkdownDescription: "Adaptive allocations configuration. When enabled, the number of allocations is set based on the current load. Cannot be set when `number_of_allocations` is configured.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "If `true`, adaptive allocations is enabled.",
						Required:            true,
					},
					"min_number_of_allocations": schema.Int64Attribute{
						MarkdownDescription: "Specifies the minimum number of allocations to scale to.",
						Optional:            true,
					},
					"max_number_of_allocations": schema.Int64Attribute{
						MarkdownDescription: "Specifies the maximum number of allocations to scale to.",
						Optional:            true,
					},
				},
			},
		},
	}
}
