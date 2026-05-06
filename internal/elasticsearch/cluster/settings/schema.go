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

package settings

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema() schema.Schema {
	settingNestedAttr := schema.SetNestedAttribute{
		MarkdownDescription: "Defines the settings in the cluster.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "The name of the setting to set and track.",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "The value of the setting to set and track.",
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"value_list": schema.ListAttribute{
					MarkdownDescription: "The list of values to be set for the key, where the list is required.",
					Optional:            true,
					Computed:            true,
					ElementType:         types.StringType,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
				},
			},
		},
	}

	settingsBlock := schema.ListNestedBlock{
		MarkdownDescription: "Settings block containing individual setting entries.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"setting": settingNestedAttr,
			},
		},
	}

	return schema.Schema{
		MarkdownDescription: settingsResourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"persistent": settingsBlock,
			"transient":  settingsBlock,
		},
	}
}
