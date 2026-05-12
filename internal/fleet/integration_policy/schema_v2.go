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

package integrationpolicy

import (
	"context"
	"os"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

// integrationPolicyModelV2 is a snapshot of the V2 model. It is retained so the
// V2→V3 state upgrader can decode prior state. It differs from the live model
// only by the presence of the now-removed `enabled` attribute.
type integrationPolicyModelV2 struct {
	ID                 types.String  `tfsdk:"id"`
	KibanaConnection   types.List    `tfsdk:"kibana_connection"`
	PolicyID           types.String  `tfsdk:"policy_id"`
	Name               types.String  `tfsdk:"name"`
	Namespace          types.String  `tfsdk:"namespace"`
	AgentPolicyID      types.String  `tfsdk:"agent_policy_id"`
	AgentPolicyIDs     types.List    `tfsdk:"agent_policy_ids"`
	Description        types.String  `tfsdk:"description"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	Force              types.Bool    `tfsdk:"force"`
	IntegrationName    types.String  `tfsdk:"integration_name"`
	IntegrationVersion types.String  `tfsdk:"integration_version"`
	OutputID           types.String  `tfsdk:"output_id"`
	Inputs             InputsValue   `tfsdk:"inputs"`
	VarsJSON           VarsJSONValue `tfsdk:"vars_json"`
	SpaceIDs           types.Set     `tfsdk:"space_ids"`
}

// getSchemaV2 returns the prior V2 resource schema. It is identical to the live
// V3 schema except for the additional top-level `enabled` attribute and the
// `Version: 2` marker. This schema is used as `PriorSchema` for the V2→V3 state
// upgrader; it is NOT returned by the resource's live `Schema` method.
func getSchemaV2() schema.Schema {
	varsAreSensitive := !logging.IsDebugOrHigher() && os.Getenv("TF_ACC") != "1"
	return schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":      schema.StringAttribute{Required: true},
			"namespace": schema.StringAttribute{Required: true},
			"agent_policy_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression()),
				},
			},
			"agent_policy_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Root("agent_policy_id").Expression()),
					listvalidator.SizeAtLeast(1),
				},
			},
			"description": schema.StringAttribute{Optional: true},
			"enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
			},
			"force":               schema.BoolAttribute{Optional: true},
			"integration_name":    schema.StringAttribute{Required: true},
			"integration_version": schema.StringAttribute{Required: true},
			"output_id":           schema.StringAttribute{Optional: true},
			"vars_json": schema.StringAttribute{
				CustomType: VarsJSONType{
					JSONWithContextualDefaultsType: customtypes.NewJSONWithContextualDefaultsType(populateVarsJSONDefaults),
				},
				Computed:  true,
				Optional:  true,
				Sensitive: varsAreSensitive,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"inputs": schema.MapNestedAttribute{
				CustomType: NewInputsType(NewInputType(getInputsAttributeTypes())),
				Computed:   true,
				Optional:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					CustomType: NewInputType(getInputsAttributeTypes()),
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Computed: true,
							Optional: true,
							Default:  booldefault.StaticBool(true),
						},
						"vars": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Sensitive:  varsAreSensitive,
						},
						"defaults": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"vars": schema.StringAttribute{
									CustomType: jsontypes.NormalizedType{},
									Computed:   true,
								},
								"streams": schema.MapNestedAttribute{
									Computed: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"enabled": schema.BoolAttribute{Computed: true},
											"vars": schema.StringAttribute{
												CustomType: jsontypes.NormalizedType{},
												Computed:   true,
											},
										},
									},
								},
							},
						},
						"streams": schema.MapNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Computed: true,
										Optional: true,
										Default:  booldefault.StaticBool(true),
									},
									"vars": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Optional:   true,
										Sensitive:  varsAreSensitive,
									},
								},
							},
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}

// upgradeV2ToV3 drops the now-removed top-level `enabled` attribute when
// migrating prior V2 state to the live V3 schema. Every other field is carried
// over verbatim. No Fleet API calls are made.
func upgradeV2ToV3(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior integrationPolicyModelV2

	diags := req.State.Get(ctx, &prior)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	next := integrationPolicyModel{
		ID:                 prior.ID,
		KibanaConnection:   prior.KibanaConnection,
		PolicyID:           prior.PolicyID,
		Name:               prior.Name,
		Namespace:          prior.Namespace,
		AgentPolicyID:      prior.AgentPolicyID,
		AgentPolicyIDs:     prior.AgentPolicyIDs,
		Description:        prior.Description,
		Force:              prior.Force,
		IntegrationName:    prior.IntegrationName,
		IntegrationVersion: prior.IntegrationVersion,
		OutputID:           prior.OutputID,
		Inputs:             prior.Inputs,
		VarsJSON:           prior.VarsJSON,
		SpaceIDs:           prior.SpaceIDs,
	}

	if prior.KibanaConnection.IsNull() || prior.KibanaConnection.IsUnknown() {
		next.KibanaConnection = providerschema.KibanaConnectionNullList()
	}

	diags = resp.State.Set(ctx, next)
	resp.Diagnostics.Append(diags...)
}
