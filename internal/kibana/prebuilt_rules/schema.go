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

package prebuiltrules

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *PrebuiltRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rules_installed": schema.Int64Attribute{
				Description: "Number of prebuilt rules that are installed.",
				Computed:    true,
			},
			"rules_not_installed": schema.Int64Attribute{
				Description: "Number of prebuilt rules that are not installed.",
				Computed:    true,
			},
			"rules_not_updated": schema.Int64Attribute{
				Description: "Number of prebuilt rules that have updates available.",
				Computed:    true,
			},
			"timelines_installed": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that are installed.",
				Computed:    true,
			},
			"timelines_not_installed": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that are not installed.",
				Computed:    true,
			},
			"timelines_not_updated": schema.Int64Attribute{
				Description: "Number of prebuilt timelines that have updates available.",
				Computed:    true,
			},
		},
	}
}
