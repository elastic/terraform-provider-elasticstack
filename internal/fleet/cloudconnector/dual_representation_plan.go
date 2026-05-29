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

package cloudconnector

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// reconcileDualRepresentationPlan copies read-populated sibling attributes from
// state into plan so Optional (non-Computed) aws/vars parents do not show
// spurious removal diffs after Read dual-populates both representations.
func reconcileDualRepresentationPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
	config cloudConnectorModel,
) {
	var state cloudConnectorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan cloudConnectorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch {
	case isNestedBlockConfigured(config.AWS):
		if !awsBlockEqualIgnoringWriteOnly(plan.AWS, state.AWS) {
			varsMap, varsDiags := planVarsMapFromAWSBlockForUpdate(ctx, plan, state)
			resp.Diagnostics.Append(varsDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrVarsMap), varsMap)...)
		}
	case isNestedBlockConfigured(config.Azure):
		if !azureBlockEqualIgnoringWriteOnly(plan.Azure, state.Azure) {
			varsMap, varsDiags := planVarsMapFromAzureBlockForUpdate(ctx, plan, state)
			resp.Diagnostics.Append(varsDiags...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(attrVarsMap), varsMap)...)
		}
	case isVarsMapConfigured(config.Vars):
		if isNestedBlockConfigured(config.AWS) || isNestedBlockConfigured(config.Azure) {
			return
		}
		if !plan.Vars.Equal(state.Vars) {
			return
		}
		switch {
		case isNestedBlockConfigured(state.AWS):
			copyStateSiblingToPlan(ctx, resp, state.AWS, path.Root(attrAWSBlock), config.AWS)
		case isNestedBlockConfigured(state.Azure):
			copyStateSiblingToPlan(ctx, resp, state.Azure, path.Root(attrAzureBlock), config.Azure)
		}
	}
}

func nullVarsElement() cloudConnectorVarsElement {
	return cloudConnectorVarsElement{
		String:      types.StringNull(),
		Number:      types.Float64Null(),
		Bool:        types.BoolNull(),
		Type:        types.StringNull(),
		Frozen:      types.BoolNull(),
		Value:       types.StringNull(),
		SecretValue: types.StringNull(),
		SecretRef:   types.ObjectNull(secretRefAttrTypes()),
	}
}

func structuredTextPlanElement(value types.String) cloudConnectorVarsElement {
	elem := nullVarsElement()
	elem.Type = types.StringValue(varsStructuredTypeText)
	elem.Value = value
	return elem
}

func structuredPasswordPlanElement() cloudConnectorVarsElement {
	elem := nullVarsElement()
	elem.Type = types.StringValue(varsStructuredTypePassword)
	return elem
}

// planVarsForCreate builds the create-plan vars map from config, leaving secret_ref
// unknown for password elements so apply consistency checks pass after secrets are stored.
func planVarsForCreate(ctx context.Context, configVars types.Map) (types.Map, diag.Diagnostics) {
	elems, diags := varsElementsFromMap(configVars)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
	}

	for key, elem := range elems {
		if elem.Type.ValueString() == varsStructuredTypePassword {
			elem.SecretRef = types.ObjectUnknown(secretRefAttrTypes())
		}
		elems[key] = elem
	}

	return varsMapFromElements(ctx, elems)
}

func passwordPlanElementWithStateSecretRef(state cloudConnectorModel, varKey string) cloudConnectorVarsElement {
	elem := structuredPasswordPlanElement()
	stateElems, diags := varsElementsFromMap(state.Vars)
	if diags.HasError() {
		return elem
	}
	if stateElem, ok := stateElems[varKey]; ok {
		elem.SecretRef = stateElem.SecretRef
	}
	return elem
}

func planVarsMapFromAWSBlockForUpdate(ctx context.Context, plan, state cloudConnectorModel) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	aws, awsDiags := awsBlockFromObject(plan.AWS)
	diags.Append(awsDiags...)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
	}

	elems := map[string]cloudConnectorVarsElement{
		attrAWSRoleArn:    structuredTextPlanElement(aws.RoleArn),
		attrAWSExternalID: passwordPlanElementWithStateSecretRef(state, attrAWSExternalID),
	}

	return varsMapFromElements(ctx, elems)
}

func planVarsMapFromAzureBlockForUpdate(ctx context.Context, plan, state cloudConnectorModel) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	azure, azureDiags := azureBlockFromObject(plan.Azure)
	diags.Append(azureDiags...)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
	}

	elems := map[string]cloudConnectorVarsElement{
		attrAzureTenantID:                       passwordPlanElementWithStateSecretRef(state, attrAzureTenantID),
		attrAzureClientID:                       passwordPlanElementWithStateSecretRef(state, attrAzureClientID),
		wireKeyAzureCredentialsCloudConnectorID: structuredTextPlanElement(azure.CloudConnectorID),
	}

	return varsMapFromElements(ctx, elems)
}

func planVarsMapFromAWSBlock(ctx context.Context, config cloudConnectorModel) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	aws, awsDiags := awsBlockFromObject(config.AWS)
	diags.Append(awsDiags...)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
	}

	elems := map[string]cloudConnectorVarsElement{
		attrAWSRoleArn:    structuredTextPlanElement(aws.RoleArn),
		attrAWSExternalID: structuredPasswordPlanElement(),
	}

	return varsMapFromElements(ctx, elems)
}

func planVarsMapFromAzureBlock(ctx context.Context, config cloudConnectorModel) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	azure, azureDiags := azureBlockFromObject(config.Azure)
	diags.Append(azureDiags...)
	if diags.HasError() {
		return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
	}

	elems := map[string]cloudConnectorVarsElement{
		attrAzureTenantID:                       structuredPasswordPlanElement(),
		attrAzureClientID:                       structuredPasswordPlanElement(),
		wireKeyAzureCredentialsCloudConnectorID: structuredTextPlanElement(azure.CloudConnectorID),
	}

	return varsMapFromElements(ctx, elems)
}

func varsMapFromElements(_ context.Context, elems map[string]cloudConnectorVarsElement) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrValues := make(map[string]attr.Value, len(elems))
	for key, elem := range elems {
		obj, objDiags := varsElementToObject(elem)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
		}
		attrValues[key] = obj
	}
	varsMap, mapDiags := types.MapValue(types.ObjectType{AttrTypes: varsElementAttrTypes()}, attrValues)
	diags.Append(mapDiags...)
	return varsMap, diags
}

func isNestedBlockConfigured(block types.Object) bool {
	return typeutils.IsKnown(block)
}

func isVarsMapConfigured(vars types.Map) bool {
	return typeutils.IsKnown(vars)
}

func copyStateSiblingToPlan(
	ctx context.Context,
	resp *resource.ModifyPlanResponse,
	stateValue attr.Value,
	planPath path.Path,
	configValue attr.Value,
) {
	if typeutils.IsKnown(configValue) {
		return
	}
	if stateValue.IsNull() {
		return
	}
	if stateValue.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Skipped dual representation plan reconciliation",
			fmt.Sprintf("Could not copy %s from state into plan because the state value is unknown.", planPath.String()),
		)
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, planPath, stateValue)...)
}
