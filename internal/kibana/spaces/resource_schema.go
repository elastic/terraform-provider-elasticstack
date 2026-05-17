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

package spaces

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Creates a Kibana space. See the [spaces API documentation](https://www.elastic.co/guide/en/kibana/master/spaces-api-post.html) for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: spaceAttrDescResourceID,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "The space ID that is part of the Kibana URL when inside the space.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9_-]+$`), "must only contain lowercase letters, numbers, hyphens, and underscores"),
				},
			},
			"name": schema.StringAttribute{
				Description: spaceAttrDescName,
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: spaceAttrDescDescription,
				Optional:    true,
			},
			"disabled_features": schema.SetAttribute{
				Description: spaceAttrDescDisabledFeatures,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"initials": schema.StringAttribute{
				Description: spaceAttrDescInitials,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"color": schema.StringAttribute{
				Description: spaceAttrDescColor,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"image_url": schema.StringAttribute{
				Description: spaceAttrDescImageURL,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^data:image/`), "must be a valid data-URL encoded image"),
				},
			},
			"solution": schema.StringAttribute{
				Description: spaceAttrDescSolution,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("security", "oblt", "es", "classic"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
