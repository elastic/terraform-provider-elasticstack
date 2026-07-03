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

	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
)

// integrationPolicyInputsModelV2 and integrationPolicyInputStreamModelV2 are
// frozen snapshots of the pre-condition input/stream models (see
// getInputsAttributeTypesV2). They exist solely so upgradeV2ToV3 can decode
// prior V2 `inputs` values, which have no `condition` key.
type integrationPolicyInputsModelV2 struct {
	Enabled  types.Bool           `tfsdk:"enabled"`
	Vars     jsontypes.Normalized `tfsdk:"vars"`
	Defaults types.Object         `tfsdk:"defaults"`
	Streams  types.Map            `tfsdk:"streams"`
}

type integrationPolicyInputStreamModelV2 struct {
	Enabled types.Bool           `tfsdk:"enabled"`
	Vars    jsontypes.Normalized `tfsdk:"vars"`
}

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
	varsAreSensitive := debugutils.IsSensitiveInSchema()
	return schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrPolicyID: schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrName:      schema.StringAttribute{Required: true},
			attrNamespace: schema.StringAttribute{Required: true},
			attrAgentPolicyID: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root(attrAgentPolicyIDs).Expression()),
				},
			},
			attrAgentPolicyIDs: schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Root(attrAgentPolicyID).Expression()),
					listvalidator.SizeAtLeast(1),
				},
			},
			attrDescription: schema.StringAttribute{Optional: true},
			attrEnabled: schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(true),
			},
			attrForce:              schema.BoolAttribute{Optional: true},
			attrIntegrationName:    schema.StringAttribute{Required: true},
			attrIntegrationVersion: schema.StringAttribute{Required: true},
			attrOutputID:           schema.StringAttribute{Optional: true},
			attrVarsJSON: schema.StringAttribute{
				CustomType: policyshape.NewVarsJSONType(lookupCachedPackageInfo),
				Computed:   true,
				Optional:   true,
				Sensitive:  varsAreSensitive,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrSpaceIDs: schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"inputs": schema.MapNestedAttribute{
				CustomType: NewInputsType(NewInputType(getInputsAttributeTypesV2())),
				Computed:   true,
				Optional:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					CustomType: NewInputType(getInputsAttributeTypesV2()),
					Attributes: map[string]schema.Attribute{
						attrEnabled: schema.BoolAttribute{
							Computed: true,
							Optional: true,
							Default:  booldefault.StaticBool(true),
						},
						attrVars: schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Sensitive:  varsAreSensitive,
						},
						attrDefaults: schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								attrVars: schema.StringAttribute{
									CustomType: jsontypes.NormalizedType{},
									Computed:   true,
								},
								attrStreams: schema.MapNestedAttribute{
									Computed: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											attrEnabled: schema.BoolAttribute{Computed: true},
											attrVars: schema.StringAttribute{
												CustomType: jsontypes.NormalizedType{},
												Computed:   true,
											},
										},
									},
								},
							},
						},
						attrStreams: schema.MapNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									attrEnabled: schema.BoolAttribute{
										Computed: true,
										Optional: true,
										Default:  booldefault.StaticBool(true),
									},
									attrVars: schema.StringAttribute{
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

// getInputsAttributeTypesV2 is a frozen snapshot of the pre-condition
// inputs-map attribute types (i.e. what getInputsAttributeTypes() returned
// before the Phase 1 `condition` addition). It MUST NOT be changed: it is
// used as PriorSchema to decode raw V2 state written before `condition`
// existed, and that raw state has no `condition` key. Delegating to the live
// (shared, condition-including) policyshape types here would make the
// PriorSchema's expected object shape not match the historical wire format,
// breaking the V2->V3 state upgrade for any resource still on schema V2.
func getInputsAttributeTypesV2() map[string]attr.Type {
	return map[string]attr.Type{
		attrEnabled: types.BoolType,
		attrVars:    jsontypes.NormalizedType{},
		attrDefaults: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				attrVars: jsontypes.NormalizedType{},
				attrStreams: types.MapType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							attrEnabled: types.BoolType,
							attrVars:    jsontypes.NormalizedType{},
						},
					},
				},
			},
		},
		attrStreams: types.MapType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					attrEnabled: types.BoolType,
					attrVars:    jsontypes.NormalizedType{},
				},
			},
		},
	}
}

// upgradeV2ToV3 drops the now-removed top-level `enabled` attribute when
// migrating prior V2 state to the live V3 schema, and rebuilds `inputs` to
// match the live (condition-including) element type. Every other field is
// carried over verbatim. No Fleet API calls are made.
func upgradeV2ToV3(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior integrationPolicyModelV2

	diags := req.State.Get(ctx, &prior)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputs, d := convertInputsV2ToV3(ctx, prior.Inputs)
	resp.Diagnostics.Append(d...)
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
		Inputs:             inputs,
		VarsJSON:           prior.VarsJSON,
		SpaceIDs:           prior.SpaceIDs,
	}

	if prior.KibanaConnection.IsNull() || prior.KibanaConnection.IsUnknown() {
		next.KibanaConnection = providerschema.KibanaConnectionNullList()
	}

	diags = resp.State.Set(ctx, next)
	resp.Diagnostics.Append(diags...)
}

// convertInputsV2ToV3 rebuilds a V2 (frozen, pre-`condition`) InputsValue
// into an InputsValue compatible with the live V3 element type, filling
// `condition` as null on every input and stream (V2 state never had it).
func convertInputsV2ToV3(ctx context.Context, prior InputsValue) (InputsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if prior.IsUnknown() {
		return InputsValue{MapValue: types.MapUnknown(getInputsElementType())}, diags
	}
	if !typeutils.IsKnown(prior) || prior.IsNull() {
		return NewInputsNull(getInputsElementType()), diags
	}

	priorMap := typeutils.MapTypeAs[integrationPolicyInputsModelV2](ctx, prior.MapValue, path.Root("inputs"), &diags)
	if diags.HasError() {
		return NewInputsNull(getInputsElementType()), diags
	}

	next := make(map[string]integrationPolicyInputsModel, len(priorMap))
	for id, in := range priorMap {
		streams, d := convertStreamsV2ToV3(ctx, in.Streams)
		diags.Append(d...)
		if diags.HasError() {
			return NewInputsNull(getInputsElementType()), diags
		}

		next[id] = integrationPolicyInputsModel{
			Enabled:   in.Enabled,
			Condition: types.StringNull(),
			Vars:      in.Vars,
			Defaults:  in.Defaults,
			Streams:   streams,
		}
	}

	return NewInputsValueFrom(ctx, getInputsElementType(), next)
}

// convertStreamsV2ToV3 is the per-input-streams-map counterpart of
// convertInputsV2ToV3.
func convertStreamsV2ToV3(ctx context.Context, streams types.Map) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if !typeutils.IsKnown(streams) || streams.IsNull() {
		return types.MapNull(getInputStreamType()), diags
	}

	priorStreams := typeutils.MapTypeAs[integrationPolicyInputStreamModelV2](ctx, streams, path.Root("streams"), &diags)
	if diags.HasError() {
		return types.MapNull(getInputStreamType()), diags
	}

	next := make(map[string]integrationPolicyInputStreamModel, len(priorStreams))
	for id, s := range priorStreams {
		next[id] = integrationPolicyInputStreamModel{
			Enabled:   s.Enabled,
			Condition: types.StringNull(),
			Vars:      s.Vars,
		}
	}

	return types.MapValueFrom(ctx, getInputStreamType(), next)
}
