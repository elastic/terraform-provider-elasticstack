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

package alertingrule

import (
	"context"
	"sync"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	attrTypesOnce        sync.Once
	cachedActionsTypes   map[string]attr.Type
	cachedFrequencyTypes map[string]attr.Type
	cachedFilterTypes    map[string]attr.Type
	cachedTimeframeTypes map[string]attr.Type
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: resourceDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the alerting rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule_id": schema.StringAttribute{
				Description: "The identifier for the rule. Until Kibana version 8.17.0 this should be a UUID v1 or v4, for later versions any format can be used. If it is omitted, an ID is randomly generated.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
			"name": schema.StringAttribute{
				Description: "The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.",
				Required:    true,
			},
			"consumer": schema.StringAttribute{
				Description: "The name of the application or feature that owns the rule.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"notify_when": schema.StringAttribute{
				Description: notifyWhenDescription,
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("onActionGroupChange", "onActiveAlert", "onThrottleInterval"),
				},
			},
			"params": schema.StringAttribute{
				Description: "The rule parameters, which differ for each rule type.",
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"rule_type_id": schema.StringAttribute{
				Description: ruleTypeIDDescription,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"interval": schema.StringAttribute{
				Description: "The check interval, which specifies how frequently the rule conditions are checked. The interval must be specified in seconds, minutes, hours or days.",
				Required:    true,
				Validators: []validator.String{
					validators.StringIsAlertingDuration{},
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Indicates if you want to run the rule on an interval basis.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"tags": schema.SetAttribute{
				Description: "A list of tag names that are applied to the rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"throttle": schema.StringAttribute{
				Description: throttleRuleDescription,
				Optional:    true,
				Validators: []validator.String{
					validators.StringIsAlertingDuration{},
				},
			},
			"scheduled_task_id": schema.StringAttribute{
				Description: "ID of the scheduled task that will execute the alert.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_execution_status": schema.StringAttribute{
				Description: "Status of the last execution of this rule.",
				Computed:    true,
			},
			"last_execution_date": schema.StringAttribute{
				Description: "Date of the last execution of this rule.",
				Computed:    true,
			},
			"alert_delay": schema.Int64Attribute{
				Description: "A number that indicates how many consecutive runs need to meet the rule conditions for an alert to occur.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"actions": schema.ListNestedBlock{
				Description: "An action that runs under defined conditions.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							Description: actionsGroupDescription,
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("default"),
						},
						"id": schema.StringAttribute{
							Description: "The identifier for the connector saved object.",
							Required:    true,
						},
						"params": schema.StringAttribute{
							Description: "The parameters for the action, which are sent to the connector.",
							Required:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
					},
					Blocks: map[string]schema.Block{
						"frequency": schema.SingleNestedBlock{
							Description: actionsFrequencyDescription,
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(path.MatchRelative().AtName("summary")),
								objectvalidator.AlsoRequires(path.MatchRelative().AtName("notify_when")),
							},
							Attributes: map[string]schema.Attribute{
								"summary": schema.BoolAttribute{
									Description: "Indicates whether the action is a summary.",
									Optional:    true,
									Computed:    true,
								},
								"notify_when": schema.StringAttribute{
									Description: actionsFrequencyNotifyWhenDescription,
									Optional:    true,
									Computed:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("onActionGroupChange", "onActiveAlert", "onThrottleInterval"),
									},
								},
								"throttle": schema.StringAttribute{
									Description: actionsFrequencyThrottleDescription,
									Optional:    true,
									Validators: []validator.String{
										validators.StringIsAlertingDuration{},
									},
								},
							},
						},
						"alerts_filter": schema.SingleNestedBlock{
							Description: alertsFilterDescription,
							Attributes: map[string]schema.Attribute{
								"kql": schema.StringAttribute{
									Description: "Defines a query filter that determines whether the action runs. Written in Kibana Query Language (KQL).",
									Optional:    true,
								},
							},
							Blocks: map[string]schema.Block{
								// The timeframe attributes (days, timezone, hours_start, hours_end) are
								// marked as Optional instead of Required so the block can be omitted entirely.
								// When the block IS present, the objectvalidator.AlsoRequires validators ensure
								// all fields are provided, as required by the Kibana API.
								"timeframe": schema.SingleNestedBlock{
									Description: "Defines a period that limits whether the action runs.",
									Validators: []validator.Object{
										objectvalidator.AlsoRequires(path.MatchRelative().AtName("days")),
										objectvalidator.AlsoRequires(path.MatchRelative().AtName("timezone")),
										objectvalidator.AlsoRequires(path.MatchRelative().AtName("hours_start")),
										objectvalidator.AlsoRequires(path.MatchRelative().AtName("hours_end")),
									},
									Attributes: map[string]schema.Attribute{
										"days": schema.ListAttribute{
											Description: timeframeDaysDescription,
											Optional:    true,
											ElementType: types.Int64Type,
											Validators: []validator.List{
												listvalidator.ValueInt64sAre(int64validator.Between(1, 7)),
											},
										},
										"timezone": schema.StringAttribute{
											Description: "The ISO time zone for the hours values. Values such as UTC and UTC+1 also work but lack built-in daylight savings time support and are not recommended.",
											Optional:    true,
										},
										"hours_start": schema.StringAttribute{
											Description: "Defines the range of time in a day that the action can run. The start of the time frame in 24-hour notation (hh:mm).",
											Optional:    true,
											Validators: []validator.String{
												validators.StringIsHours{},
											},
										},
										"hours_end": schema.StringAttribute{
											Description: "Defines the range of time in a day that the action can run. The end of the time frame in 24-hour notation (hh:mm).",
											Optional:    true,
											Validators: []validator.String{
												validators.StringIsHours{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// initAttrTypes initializes and caches all attribute types from the schema.
// This is called once via sync.Once to avoid repeatedly parsing the schema.
func initAttrTypes() {
	s := getSchema()

	actionsBlock := s.Blocks["actions"].(schema.ListNestedBlock)
	cachedActionsTypes = actionsBlock.NestedObject.Type().(attr.TypeWithAttributeTypes).AttributeTypes()

	freqBlock := actionsBlock.NestedObject.Blocks["frequency"].(schema.SingleNestedBlock)
	cachedFrequencyTypes = freqBlock.Type().(attr.TypeWithAttributeTypes).AttributeTypes()

	filterBlock := actionsBlock.NestedObject.Blocks["alerts_filter"].(schema.SingleNestedBlock)
	cachedFilterTypes = filterBlock.Type().(attr.TypeWithAttributeTypes).AttributeTypes()

	tfBlock := filterBlock.Blocks["timeframe"].(schema.SingleNestedBlock)
	cachedTimeframeTypes = tfBlock.Type().(attr.TypeWithAttributeTypes).AttributeTypes()
}

// getActionsAttrTypes returns the attribute types for actions list elements.
func getActionsAttrTypes() map[string]attr.Type {
	attrTypesOnce.Do(initAttrTypes)
	return cachedActionsTypes
}

// getFrequencyAttrTypes returns the attribute types for frequency object.
func getFrequencyAttrTypes() map[string]attr.Type {
	attrTypesOnce.Do(initAttrTypes)
	return cachedFrequencyTypes
}

// getAlertsFilterAttrTypes returns the attribute types for alerts_filter object.
func getAlertsFilterAttrTypes() map[string]attr.Type {
	attrTypesOnce.Do(initAttrTypes)
	return cachedFilterTypes
}

// getTimeframeAttrTypes returns the attribute types for timeframe object.
func getTimeframeAttrTypes() map[string]attr.Type {
	attrTypesOnce.Do(initAttrTypes)
	return cachedTimeframeTypes
}
