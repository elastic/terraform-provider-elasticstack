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

package datafeedstate

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

//go:embed resource-description.md
var description string

func GetSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"datafeed_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the ML datafeed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The desired state for the ML datafeed. Valid values are `started` and `stopped`.",
				Required:            true,
				Validators: []validator.String{
					// We don't allow starting/stopping here since they're transient states
					stringvalidator.OneOf(string(datafeed.StateStarted), string(datafeed.StateStopped)),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "When stopping a datafeed, use to forcefully stop it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"start": schema.StringAttribute{
				MarkdownDescription: "The time that the datafeed should start collecting data. " +
					"When not specified, the datafeed starts in real-time. " +
					"This value is preserved verbatim in state and is passed to the Start Datafeed API; " +
					"Elasticsearch may use a different effective search start, which is reported in `effective_search_start`. " +
					"This property must be specified in RFC 3339 format.",
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
			},
			"end": schema.StringAttribute{
				MarkdownDescription: "The time that the datafeed should end collecting data. " +
					"When not specified, the datafeed continues in real-time. " +
					"This value is preserved verbatim in state and is passed to the Start Datafeed API; " +
					"Elasticsearch may use a different effective search end, which is reported in `effective_search_end`. " +
					"This property must be specified in RFC 3339 format.",
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
			},
			"effective_search_start": schema.StringAttribute{
				MarkdownDescription: "The effective search start time reported by Elasticsearch for a started datafeed " +
					"(`running_state.search_interval.start_ms`). " +
					"Null when the datafeed is stopped or when `running_state` / `search_interval` is absent.",
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"effective_search_end": schema.StringAttribute{
				MarkdownDescription: "The effective search end time reported by Elasticsearch for a started datafeed " +
					"(`running_state.search_interval.end_ms`). " +
					"Null when the datafeed is stopped, when `running_state.real_time_configured` is true, " +
					"or when `running_state` / `search_interval` is absent.",
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"datafeed_timeout": schema.StringAttribute{
				MarkdownDescription: "Timeout for the operation. Examples: `30s`, `5m`, `1h`. Default is `30s`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("30s"),
				CustomType:          customtypes.DurationType{},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}
