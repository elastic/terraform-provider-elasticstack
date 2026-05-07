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

package watch

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// watchSchema returns the schema for the watch resource without the
// elasticsearch_connection block; the envelope injects it.
func watchSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manage Watches. See the [Watcher API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api.html) for more details.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"watch_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the watch.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Defines whether the watch is active or inactive by default. The default value is true, which means the watch is active by default.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "The trigger that defines when the watch should run.",
				Required:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"input": schema.StringAttribute{
				MarkdownDescription: "The input that defines the input that loads the data for the watch.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				Default:             stringdefault.StaticString(`{"none":{}}`),
			},
			"condition": schema.StringAttribute{
				MarkdownDescription: "The condition that defines if the actions should be run.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				Default:             stringdefault.StaticString(`{"always":{}}`),
			},
			"actions": schema.StringAttribute{
				MarkdownDescription: "The list of actions that will be run if the condition matches.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				Default:             stringdefault.StaticString(`{}`),
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Metadata json that will be copied into the history entries.",
				Optional:            true,
				Computed:            true,
				CustomType:          jsontypes.NormalizedType{},
				Default:             stringdefault.StaticString(`{}`),
			},
			"transform": schema.StringAttribute{
				MarkdownDescription: "Processes the watch payload to prepare it for the watch actions.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"throttle_period_in_millis": schema.Int64Attribute{
				MarkdownDescription: "Minimum time in milliseconds between actions being run. Defaults to 5000.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(5000),
			},
		},
	}
}
