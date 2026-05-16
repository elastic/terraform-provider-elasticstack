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
	_ "embed" // Used for embedding schema descriptions

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationPolicyModelV1 struct {
	ID                 types.String         `tfsdk:"id"`
	PolicyID           types.String         `tfsdk:"policy_id"`
	Name               types.String         `tfsdk:"name"`
	Namespace          types.String         `tfsdk:"namespace"`
	AgentPolicyID      types.String         `tfsdk:"agent_policy_id"`
	AgentPolicyIDs     types.List           `tfsdk:"agent_policy_ids"`
	Description        types.String         `tfsdk:"description"`
	Enabled            types.Bool           `tfsdk:"enabled"`
	Force              types.Bool           `tfsdk:"force"`
	IntegrationName    types.String         `tfsdk:"integration_name"`
	IntegrationVersion types.String         `tfsdk:"integration_version"`
	OutputID           types.String         `tfsdk:"output_id"`
	Input              types.List           `tfsdk:"input"` // > integrationPolicyInputModel
	VarsJSON           jsontypes.Normalized `tfsdk:"vars_json"`
	SpaceIDs           types.Set            `tfsdk:"space_ids"`
}

type integrationPolicyInputModelV1 struct {
	InputID     types.String         `tfsdk:"input_id"`
	Enabled     types.Bool           `tfsdk:"enabled"`
	StreamsJSON jsontypes.Normalized `tfsdk:"streams_json"`
	VarsJSON    jsontypes.Normalized `tfsdk:"vars_json"`
}

func (m integrationPolicyModelV1) toV3(ctx context.Context) (integrationPolicyModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	var varsJSONVal VarsJSONValue
	switch {
	case m.VarsJSON.IsNull():
		varsJSONVal = NewVarsJSONNull()
	case m.VarsJSON.IsUnknown():
		varsJSONVal = NewVarsJSONUnknown()
	default:
		var d diag.Diagnostics
		varsJSONVal, d = NewVarsJSONWithIntegration(m.VarsJSON.ValueString(), m.IntegrationName.ValueString(), m.IntegrationVersion.ValueString())
		diags.Append(d...)
	}

	// Convert V1 model to live (V3) model. The legacy top-level `enabled`
	// attribute is deliberately dropped: the Kibana Fleet package-policy
	// request API does not accept it, so it has never had any effect on the
	// wire and was removed from the schema in V3.
	stateModelV3 := integrationPolicyModel{
		ID:                 m.ID,
		KibanaConnection:   providerschema.KibanaConnectionNullList(),
		PolicyID:           m.PolicyID,
		Name:               m.Name,
		Namespace:          m.Namespace,
		AgentPolicyID:      m.AgentPolicyID,
		AgentPolicyIDs:     m.AgentPolicyIDs,
		Description:        m.Description,
		Force:              m.Force,
		IntegrationName:    m.IntegrationName,
		IntegrationVersion: m.IntegrationVersion,
		OutputID:           m.OutputID,
		SpaceIDs:           m.SpaceIDs,
		VarsJSON:           varsJSONVal,
	}

	inputsV1 := typeutils.ListTypeAs[integrationPolicyInputModelV1](ctx, m.Input, path.Root("input"), &diags)
	inputsV3 := make(map[string]integrationPolicyInputsModel, len(inputsV1))

	for _, inputV1 := range inputsV1 {
		id := inputV1.InputID.ValueString()
		streams, d := updateStreamsV1ToV2(ctx, inputV1.StreamsJSON, id)
		diags.Append(d...)
		if diags.HasError() {
			return stateModelV3, diags
		}

		inputsV3[id] = integrationPolicyInputsModel{
			Enabled:  inputV1.Enabled,
			Vars:     inputV1.VarsJSON,
			Streams:  streams,
			Defaults: types.ObjectNull(getInputDefaultsAttrTypes()),
		}
	}

	inputsValue, d := NewInputsValueFrom(ctx, getInputsElementType(), inputsV3)
	diags.Append(d...)

	stateModelV3.Inputs = inputsValue
	return stateModelV3, diags
}

// upgradeV1ToV3 upgrades V1 state directly to the live V3 schema. V1 used a list
// `input` block and JSON-string stream payloads; V3 uses an `inputs` map with a
// nested `streams` map (see updateStreamsV1ToV2 for the shared list-to-map logic
// retained from the V1→V2 implementation). The legacy top-level `enabled`
// attribute present in V1/V2 is dropped during the conversion.
func upgradeV1ToV3(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var stateModelV1 integrationPolicyModelV1

	diags := req.State.Get(ctx, &stateModelV1)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModelV3, diags := stateModelV1.toV3(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, stateModelV3)
	resp.Diagnostics.Append(diags...)
}

func updateStreamsV1ToV2(ctx context.Context, v1 jsontypes.Normalized, inputID string) (types.Map, diag.Diagnostics) {
	if !typeutils.IsKnown(v1) {
		return types.MapNull(getInputStreamType()), nil
	}

	var apiStreams map[string]kbapi.PackagePolicyMappedInputStream
	diags := v1.Unmarshal(&apiStreams)
	if diags.HasError() {
		return types.MapNull(getInputStreamType()), diags
	}

	if len(apiStreams) == 0 {
		return types.MapNull(getInputStreamType()), nil
	}

	streams := make(map[string]integrationPolicyInputStreamModel)
	for streamID, streamData := range apiStreams {
		streamModel := integrationPolicyInputStreamModel{
			Enabled: types.BoolPointerValue(streamData.Enabled),
			Vars:    typeutils.MapToNormalizedType(typeutils.Deref(streamData.Vars), path.Root("inputs").AtMapKey(inputID).AtName("streams").AtMapKey(streamID).AtName("vars"), &diags),
		}

		streams[streamID] = streamModel
	}

	return types.MapValueFrom(ctx, getInputStreamType(), streams)
}

func getSchemaV1() *schema.Schema {
	return &schema.Schema{
		Version:     1,
		Description: integrationPolicyDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Unique identifier of the integration policy.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the integration policy.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the integration policy.",
				Required:    true,
			},
			"agent_policy_id": schema.StringAttribute{
				Description: "ID of the agent policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression()),
				},
			},
			"agent_policy_ids": schema.ListAttribute{
				Description: "List of agent policy IDs.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Root("agent_policy_id").Expression()),
					listvalidator.SizeAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the integration policy.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the integration policy.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force": schema.BoolAttribute{
				Description: "Force operations, such as creation and deletion, to occur.",
				Optional:    true,
			},
			"integration_name": schema.StringAttribute{
				Description: "The name of the integration package.",
				Required:    true,
			},
			"integration_version": schema.StringAttribute{
				Description: "The version of the integration package.",
				Required:    true,
			},
			"output_id": schema.StringAttribute{
				Description: "The ID of the output to send data to. When not specified, the default output of the agent policy will be used.",
				Optional:    true,
			},
			"vars_json": schema.StringAttribute{
				Description: "Integration-level variables as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
				Optional:    true,
				Sensitive:   true,
			},
			"space_ids": schema.SetAttribute{
				Description: spaceIDsDescription,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"input": schema.ListNestedBlock{
				Description: "Integration inputs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"input_id": schema.StringAttribute{
							Description: "The identifier of the input.",
							Required:    true,
						},
						"enabled": schema.BoolAttribute{
							Description: "Enable the input.",
							Computed:    true,
							Optional:    true,
							Default:     booldefault.StaticBool(true),
						},
						"streams_json": schema.StringAttribute{
							Description: "Input streams as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Computed:    true,
							Optional:    true,
							Sensitive:   true,
						},
						"vars_json": schema.StringAttribute{
							Description: "Input variables as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Computed:    true,
							Optional:    true,
							Sensitive:   true,
						},
					},
				},
			},
		},
	}
}

func getInputTypeV1() attr.Type {
	return getSchemaV1().Blocks["input"].Type().(attr.TypeWithElementType).ElementType()
}
