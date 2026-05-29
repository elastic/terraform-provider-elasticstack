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
	"encoding/json"
	"fmt"
	"maps"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type postWireVars map[string]kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

func (plan cloudConnectorModel) toAPICreateBody(config cloudConnectorModel) (kbapi.PostFleetCloudConnectorsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PostFleetCloudConnectorsJSONRequestBody{
		Name:          plan.Name.ValueString(),
		CloudProvider: kbapi.PostFleetCloudConnectorsJSONBodyCloudProvider(plan.CloudProvider.ValueString()),
	}

	if !plan.AccountType.IsNull() && !plan.AccountType.IsUnknown() {
		at := kbapi.PostFleetCloudConnectorsJSONBodyAccountType(plan.AccountType.ValueString())
		body.AccountType = &at
	}

	vars, varsDiags := plan.compileVarsForWrite(config, nil, false, nil)
	diags.Append(varsDiags...)
	if diags.HasError() {
		return body, diags
	}
	body.Vars = vars

	return body, diags
}

func (plan cloudConnectorModel) toAPIUpdateBody(
	config cloudConnectorModel,
	prior cloudConnectorModel,
	resubmitWriteOnly map[string]struct{},
) (kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody, diag.Diagnostics) {
	return plan.buildAPIUpdateBodyFromVars(config, prior, nil, resubmitWriteOnly)
}

func (plan cloudConnectorModel) buildAPIUpdateBody(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	config cloudConnectorModel,
	prior cloudConnectorModel,
	resubmitWriteOnly map[string]struct{},
) (kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody, diag.Diagnostics) {
	vars, varsDiags := plan.compileVarsForWrite(config, &prior, true, resubmitWriteOnly)
	if varsDiags.HasError() {
		return kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody{}, varsDiags
	}

	resolvedVars, secretDiags := resolveWireVarsSecrets(
		ctx,
		client.GetElasticsearchTypedClient(),
		plan.CloudProvider.ValueString(),
		vars,
	)
	if secretDiags.HasError() {
		return kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody{}, secretDiags
	}

	return plan.buildAPIUpdateBodyFromVars(config, prior, resolvedVars, resubmitWriteOnly)
}

func (plan cloudConnectorModel) buildAPIUpdateBodyFromVars(
	config cloudConnectorModel,
	prior cloudConnectorModel,
	resolvedVars postWireVars,
	resubmitWriteOnly map[string]struct{},
) (kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := kbapi.PutFleetCloudConnectorsCloudconnectoridJSONRequestBody{
		Name: plan.Name.ValueStringPointer(),
	}

	if !plan.AccountType.IsNull() && !plan.AccountType.IsUnknown() {
		at := kbapi.PutFleetCloudConnectorsCloudconnectoridJSONBodyAccountType(plan.AccountType.ValueString())
		body.AccountType = &at
	}

	vars := resolvedVars
	if vars == nil {
		var varsDiags diag.Diagnostics
		vars, varsDiags = plan.compileVarsForWrite(config, &prior, true, resubmitWriteOnly)
		diags.Append(varsDiags...)
		if diags.HasError() {
			return body, diags
		}
	}

	putVars, convertDiags := postWireVarsToPut(vars)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return body, diags
	}
	body.Vars = &putVars

	return body, diags
}

// compileVarsForWrite chooses the wire representation from config (user HCL only)
// and compiles field values from plan (config plus computed state from Read).
func (plan cloudConnectorModel) compileVarsForWrite(
	config cloudConnectorModel,
	prior *cloudConnectorModel,
	forUpdate bool,
	resubmitWriteOnly map[string]struct{},
) (postWireVars, diag.Diagnostics) {
	switch {
	case !config.AWS.IsNull() && !config.AWS.IsUnknown():
		aws, awsDiags := awsBlockForWrite(plan, config)
		if awsDiags.HasError() {
			return nil, awsDiags
		}
		var priorAWS *awsBlockModel
		if forUpdate && prior != nil && !prior.AWS.IsNull() && !prior.AWS.IsUnknown() {
			priorBlock, priorDiags := awsBlockFromObject(prior.AWS)
			if priorDiags.HasError() {
				return nil, priorDiags
			}
			priorAWS = &priorBlock
		}
		return compileAWS(aws, priorAWS, forUpdate, resubmitWriteOnly)
	case !config.Azure.IsNull() && !config.Azure.IsUnknown():
		azure, azureDiags := azureBlockForWrite(plan, config)
		if azureDiags.HasError() {
			return nil, azureDiags
		}
		var priorAzure *azureBlockModel
		if forUpdate && prior != nil && !prior.Azure.IsNull() && !prior.Azure.IsUnknown() {
			priorBlock, priorDiags := azureBlockFromObject(prior.Azure)
			if priorDiags.HasError() {
				return nil, priorDiags
			}
			priorAzure = &priorBlock
		}
		return compileAzure(azure, priorAzure, forUpdate, resubmitWriteOnly)
	default:
		var priorVars *types.Map
		if forUpdate && prior != nil {
			priorVars = &prior.Vars
		}
		return compileVars(varsMapForWrite(plan.Vars, config.Vars), priorVars, forUpdate)
	}
}

func awsBlockForWrite(plan, config cloudConnectorModel) (awsBlockModel, diag.Diagnostics) {
	aws, diags := awsBlockFromObject(plan.AWS)
	if diags.HasError() {
		return aws, diags
	}
	if config.AWS.IsNull() || config.AWS.IsUnknown() {
		return aws, diags
	}

	cfgAWS, cfgDiags := awsBlockFromObject(config.AWS)
	diags.Append(cfgDiags...)
	if diags.HasError() {
		return aws, diags
	}

	if typeutils.IsKnown(cfgAWS.ExternalID) {
		aws.ExternalID = cfgAWS.ExternalID
	}
	if typeutils.IsKnown(cfgAWS.RoleArn) {
		aws.RoleArn = cfgAWS.RoleArn
	}

	return aws, diags
}

func azureBlockForWrite(plan, config cloudConnectorModel) (azureBlockModel, diag.Diagnostics) {
	azure, diags := azureBlockFromObject(plan.Azure)
	if diags.HasError() {
		return azure, diags
	}
	if config.Azure.IsNull() || config.Azure.IsUnknown() {
		return azure, diags
	}

	cfgAzure, cfgDiags := azureBlockFromObject(config.Azure)
	diags.Append(cfgDiags...)
	if diags.HasError() {
		return azure, diags
	}

	if typeutils.IsKnown(cfgAzure.TenantID) {
		azure.TenantID = cfgAzure.TenantID
	}
	if typeutils.IsKnown(cfgAzure.ClientID) {
		azure.ClientID = cfgAzure.ClientID
	}
	if typeutils.IsKnown(cfgAzure.CloudConnectorID) {
		azure.CloudConnectorID = cfgAzure.CloudConnectorID
	}

	return azure, diags
}

func varsMapForWrite(plan, config types.Map) types.Map {
	if config.IsNull() || config.IsUnknown() {
		return plan
	}
	if plan.IsNull() || plan.IsUnknown() {
		return config
	}

	planElems, planDiags := varsElementsFromMap(plan)
	if planDiags.HasError() {
		return plan
	}
	configElems, configDiags := varsElementsFromMap(config)
	if configDiags.HasError() {
		return plan
	}

	merged := make(map[string]cloudConnectorVarsElement, len(planElems))
	maps.Copy(merged, planElems)
	for key, cfgElem := range configElems {
		elem := merged[key]
		if typeutils.IsKnown(cfgElem.SecretValue) {
			elem.SecretValue = cfgElem.SecretValue
		}
		if typeutils.IsKnown(cfgElem.String) {
			elem.String = cfgElem.String
		}
		if typeutils.IsKnown(cfgElem.Number) {
			elem.Number = cfgElem.Number
		}
		if typeutils.IsKnown(cfgElem.Bool) {
			elem.Bool = cfgElem.Bool
		}
		if typeutils.IsKnown(cfgElem.Type) {
			elem.Type = cfgElem.Type
		}
		if typeutils.IsKnown(cfgElem.Frozen) {
			elem.Frozen = cfgElem.Frozen
		}
		if typeutils.IsKnown(cfgElem.Value) {
			elem.Value = cfgElem.Value
		}
		merged[key] = elem
	}

	out, diags := types.MapValue(types.ObjectType{AttrTypes: varsElementAttrTypes()}, mapFromVarsElements(merged))
	if diags.HasError() {
		return plan
	}
	return out
}

func mapFromVarsElements(elems map[string]cloudConnectorVarsElement) map[string]attr.Value {
	out := make(map[string]attr.Value, len(elems))
	for key, elem := range elems {
		obj, diags := varsElementToObject(elem)
		if diags.HasError() {
			continue
		}
		out[key] = obj
	}
	return out
}

func compileAWS(aws awsBlockModel, priorAWS *awsBlockModel, forUpdate bool, resubmitWriteOnly map[string]struct{}) (postWireVars, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(postWireVars, 2)

	roleArnVar, roleDiags := wireStructuredStringPost("text", aws.RoleArn.ValueString(), nil)
	diags.Append(roleDiags...)
	if diags.HasError() {
		return nil, diags
	}
	out[attrAWSRoleArn] = roleArnVar

	externalIDVar, externalDiags := compileAWSExternalID(aws, priorAWS, forUpdate, resubmitWriteOnly)
	diags.Append(externalDiags...)
	if diags.HasError() {
		return nil, diags
	}
	out[attrAWSExternalID] = externalIDVar

	return out, diags
}

func compileAWSExternalID(
	aws awsBlockModel,
	priorAWS *awsBlockModel,
	forUpdate bool,
	resubmitWriteOnly map[string]struct{},
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	if forUpdate && priorAWS != nil {
		if ref, ok := secretRefFromObject(priorAWS.ExternalIDSecretRef); ok {
			if _, rotate := resubmitWriteOnly[writeOnlyAttributeAWSExternalID]; !rotate {
				return wireStructuredSecretRefPost("password", ref)
			}
		}
	}

	if !aws.ExternalID.IsNull() && !aws.ExternalID.IsUnknown() {
		return wireStructuredStringPost("password", aws.ExternalID.ValueString(), nil)
	}

	diags.AddError(
		"Missing AWS external_id",
		`The aws block requires external_id on create, or an existing external_id_secret_ref in state when external_id is not re-supplied on update.`,
	)
	return zero, diags
}

func compileAzure(
	azure azureBlockModel,
	priorAzure *azureBlockModel,
	forUpdate bool,
	resubmitWriteOnly map[string]struct{},
) (postWireVars, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(postWireVars, 3)

	tenantVar, tenantDiags := compileAzurePasswordVar(
		azure.TenantID,
		writeOnlyAttributeAzureTenantID,
		priorAzure,
		forUpdate,
		resubmitWriteOnly,
		func(prior *azureBlockModel) (cloudConnectorSecretRef, bool) {
			return secretRefFromObject(prior.TenantIDSecretRef)
		},
	)
	diags.Append(tenantDiags...)
	if diags.HasError() {
		return nil, diags
	}
	out[attrAzureTenantID] = tenantVar

	clientVar, clientDiags := compileAzurePasswordVar(
		azure.ClientID,
		writeOnlyAttributeAzureClientID,
		priorAzure,
		forUpdate,
		resubmitWriteOnly,
		func(prior *azureBlockModel) (cloudConnectorSecretRef, bool) {
			return secretRefFromObject(prior.ClientIDSecretRef)
		},
	)
	diags.Append(clientDiags...)
	if diags.HasError() {
		return nil, diags
	}
	out[attrAzureClientID] = clientVar

	connectorVar, connectorDiags := wireStructuredStringPost("text", azure.CloudConnectorID.ValueString(), nil)
	diags.Append(connectorDiags...)
	if diags.HasError() {
		return nil, diags
	}
	out[wireKeyAzureCredentialsCloudConnectorID] = connectorVar

	return out, diags
}

func compileAzurePasswordVar(
	value types.String,
	writeOnlyAttribute string,
	priorAzure *azureBlockModel,
	forUpdate bool,
	resubmitWriteOnly map[string]struct{},
	priorRef func(*azureBlockModel) (cloudConnectorSecretRef, bool),
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	if forUpdate && priorAzure != nil {
		if ref, ok := priorRef(priorAzure); ok {
			if _, rotate := resubmitWriteOnly[writeOnlyAttribute]; !rotate {
				return wireStructuredSecretRefPost("password", ref)
			}
		}
	}

	if !value.IsNull() && !value.IsUnknown() {
		return wireStructuredStringPost("password", value.ValueString(), nil)
	}

	diags.AddError(
		"Missing Azure secret value",
		fmt.Sprintf("The azure block requires %s on create, or an existing secret reference in state when %s is not re-supplied on update.", writeOnlyAttribute, writeOnlyAttribute),
	)
	return zero, diags
}

func compileVars(plan types.Map, prior *types.Map, forUpdate bool) (postWireVars, diag.Diagnostics) {
	var diags diag.Diagnostics

	if plan.IsNull() || plan.IsUnknown() {
		diags.AddError("Missing cloud connector vars", "The vars map must be set when aws and azure blocks are not configured.")
		return nil, diags
	}

	priorElements := map[string]cloudConnectorVarsElement{}
	if forUpdate && prior != nil && !prior.IsNull() && !prior.IsUnknown() {
		var priorDiags diag.Diagnostics
		priorElements, priorDiags = varsElementsFromMap(*prior)
		diags.Append(priorDiags...)
		if diags.HasError() {
			return nil, diags
		}
	}

	planElements, planDiags := varsElementsFromMap(plan)
	diags.Append(planDiags...)
	if diags.HasError() {
		return nil, diags
	}

	out := make(postWireVars, len(planElements))
	for key, elem := range planElements {
		var priorElem *cloudConnectorVarsElement
		if priorEntry, ok := priorElements[key]; ok {
			priorCopy := priorEntry
			priorElem = &priorCopy
		}

		wireVar, elemDiags := compileVarsElement(elem, priorElem, forUpdate, key)
		diags.Append(elemDiags...)
		if diags.HasError() {
			return nil, diags
		}
		out[key] = wireVar
	}

	return out, diags
}

func compileVarsElement(
	elem cloudConnectorVarsElement,
	prior *cloudConnectorVarsElement,
	forUpdate bool,
	key string,
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	switch {
	case !elem.String.IsNull() && !elem.String.IsUnknown():
		return wireStringPost(elem.String.ValueString())
	case !elem.Number.IsNull() && !elem.Number.IsUnknown():
		return wireNumberPost(float64ToFloat32(elem.Number.ValueFloat64()))
	case !elem.Bool.IsNull() && !elem.Bool.IsUnknown():
		return wireBoolPost(elem.Bool.ValueBool())
	case !elem.Type.IsNull() && !elem.Type.IsUnknown():
		return compileStructuredVarElement(elem, prior, forUpdate, key)
	case forUpdate && prior != nil && !prior.Type.IsNull() && !prior.Type.IsUnknown():
		return compileStructuredVarElement(elem, prior, forUpdate, key)
	default:
		diags.AddError(
			"Invalid cloud connector var",
			fmt.Sprintf("Cloud connector var %q has no supported value arm set.", key),
		)
		return zero, diags
	}
}

func compileStructuredVarElement(
	elem cloudConnectorVarsElement,
	prior *cloudConnectorVarsElement,
	forUpdate bool,
	key string,
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	var frozen *bool
	if !elem.Frozen.IsNull() && !elem.Frozen.IsUnknown() {
		frozenVal := elem.Frozen.ValueBool()
		frozen = &frozenVal
	}

	typeVal := structuredVarType(elem, prior)

	switch {
	case !elem.SecretValue.IsNull() && !elem.SecretValue.IsUnknown():
		return wireStructuredStringPost(typeVal, elem.SecretValue.ValueString(), frozen)
	case !elem.Value.IsNull() && !elem.Value.IsUnknown():
		return wireStructuredStringPost(typeVal, elem.Value.ValueString(), frozen)
	case forUpdate && prior != nil:
		if ref, ok := secretRefFromObject(prior.SecretRef); ok {
			return wireStructuredSecretRefPost(typeVal, ref)
		}
	}

	diags.AddError(
		"Missing cloud connector var value",
		fmt.Sprintf("Cloud connector var %q requires value or secret_value, or an existing secret_ref in state when the secret is not re-supplied on update.", key),
	)
	return zero, diags
}

func wireStringPost(value string) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var out kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties
	if err := out.FromPostFleetCloudConnectorsJSONBodyVars0(value); err != nil {
		return out, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to encode cloud connector var", err.Error())}
	}
	return out, nil
}

func wireNumberPost(value float32) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var out kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties
	if err := out.FromPostFleetCloudConnectorsJSONBodyVars1(value); err != nil {
		return out, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to encode cloud connector var", err.Error())}
	}
	return out, nil
}

func wireBoolPost(value bool) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var out kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties
	if err := out.FromPostFleetCloudConnectorsJSONBodyVars2(value); err != nil {
		return out, diag.Diagnostics{diag.NewErrorDiagnostic("Failed to encode cloud connector var", err.Error())}
	}
	return out, nil
}

func wireStructuredStringPost(
	typeVal, value string,
	frozen *bool,
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	structured := kbapi.PostFleetCloudConnectorsJSONBodyVars3{
		Type:   typeVal,
		Frozen: frozen,
	}
	if err := structured.Value.FromPostFleetCloudConnectorsJSONBodyVars3Value0(value); err != nil {
		diags.AddError("Failed to encode cloud connector var", err.Error())
		return out, diags
	}
	if err := out.FromPostFleetCloudConnectorsJSONBodyVars3(structured); err != nil {
		diags.AddError("Failed to encode cloud connector var", err.Error())
	}
	return out, diags
}

func wireStructuredSecretRefPost(
	typeVal string,
	ref cloudConnectorSecretRef,
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties

	structured := kbapi.PostFleetCloudConnectorsJSONBodyVars3{
		Type: typeVal,
	}
	secretRef := kbapi.PostFleetCloudConnectorsJSONBodyVars3Value1{
		Id:          ref.ID.ValueString(),
		IsSecretRef: ref.IsSecretRef.ValueBool(),
	}
	if err := structured.Value.FromPostFleetCloudConnectorsJSONBodyVars3Value1(secretRef); err != nil {
		diags.AddError("Failed to encode cloud connector secret reference", err.Error())
		return out, diags
	}
	if err := out.FromPostFleetCloudConnectorsJSONBodyVars3(structured); err != nil {
		diags.AddError("Failed to encode cloud connector var", err.Error())
	}
	return out, diags
}

func structuredVarType(elem cloudConnectorVarsElement, prior *cloudConnectorVarsElement) string {
	if !elem.Type.IsNull() && !elem.Type.IsUnknown() {
		return elem.Type.ValueString()
	}
	if prior != nil && !prior.Type.IsNull() && !prior.Type.IsUnknown() {
		return prior.Type.ValueString()
	}
	return ""
}

// postWireVarsToPut converts POST wire vars to PUT wire vars via JSON roundtrip.
// POST and PUT union types differ in Go but share the same JSON shape, so marshal
// then unmarshal avoids duplicating From* encoders. Integer-shaped JSON numbers
// remain float32 on the PUT side; empty strings and null arms roundtrip as encoded.
func postWireVarsToPut(in postWireVars) (map[string]kbapi.PutFleetCloudConnectorsCloudconnectoridJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	var diags diag.Diagnostics
	out := make(map[string]kbapi.PutFleetCloudConnectorsCloudconnectoridJSONBody_Vars_AdditionalProperties, len(in))

	for key, postVar := range in {
		raw, err := json.Marshal(postVar)
		if err != nil {
			diags.AddError("Failed to encode cloud connector var", fmt.Sprintf("Could not marshal var %q for update: %s", key, err))
			return nil, diags
		}

		var putVar kbapi.PutFleetCloudConnectorsCloudconnectoridJSONBody_Vars_AdditionalProperties
		if err := json.Unmarshal(raw, &putVar); err != nil {
			diags.AddError("Failed to encode cloud connector var", fmt.Sprintf("Could not convert var %q for update: %s", key, err))
			return nil, diags
		}
		out[key] = putVar
	}

	return out, diags
}

// float64ToFloat32 converts Terraform float64 var values to the API float32 wire type.
// Values outside float32 range or with more precision than float32 can represent may truncate.
func float64ToFloat32(v float64) float32 {
	return float32(v)
}

func awsBlockFromObject(obj types.Object) (awsBlockModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		diags.AddError("Missing AWS block", "The aws block must be set.")
		return awsBlockModel{}, diags
	}

	attrs := obj.Attributes()
	return awsBlockModel{
		RoleArn:             attrs[attrAWSRoleArn].(types.String),
		ExternalID:          attrs[attrAWSExternalID].(types.String),
		ExternalIDSecretRef: attrs[attrAWSExternalIDSecretRef].(types.Object),
	}, diags
}

func azureBlockFromObject(obj types.Object) (azureBlockModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		diags.AddError("Missing Azure block", "The azure block must be set.")
		return azureBlockModel{}, diags
	}

	attrs := obj.Attributes()
	return azureBlockModel{
		TenantID:          attrs[attrAzureTenantID].(types.String),
		ClientID:          attrs[attrAzureClientID].(types.String),
		TenantIDSecretRef: attrs[attrAzureTenantIDSecretRef].(types.Object),
		ClientIDSecretRef: attrs[attrAzureClientIDSecretRef].(types.Object),
		CloudConnectorID:  attrs[attrAzureCloudConnectorID].(types.String),
	}, diags
}

func varsElementsFromMap(vars types.Map) (map[string]cloudConnectorVarsElement, diag.Diagnostics) {
	var diags diag.Diagnostics
	elements := make(map[string]cloudConnectorVarsElement, len(vars.Elements()))

	for key, value := range vars.Elements() {
		obj, ok := value.(types.Object)
		if !ok {
			diags.AddError(
				"Invalid cloud connector vars map",
				fmt.Sprintf("Cloud connector var %q is not an object.", key),
			)
			return nil, diags
		}

		attrs := obj.Attributes()
		elements[key] = cloudConnectorVarsElement{
			String:      attrs[attrVarsString].(types.String),
			Number:      attrs[attrVarsNumber].(types.Float64),
			Bool:        attrs[attrVarsBool].(types.Bool),
			Type:        attrs[attrVarsType].(types.String),
			Frozen:      attrs[attrVarsFrozen].(types.Bool),
			Value:       attrs[attrVarsValue].(types.String),
			SecretValue: attrs[attrVarsSecretValue].(types.String),
			SecretRef:   attrs[attrVarsSecretRef].(types.Object),
		}
	}

	return elements, diags
}

func secretRefFromObject(obj types.Object) (cloudConnectorSecretRef, bool) {
	if obj.IsNull() || obj.IsUnknown() {
		return cloudConnectorSecretRef{}, false
	}

	attrs := obj.Attributes()
	id, ok := attrs[attrSecretRefID].(types.String)
	if !ok || id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		return cloudConnectorSecretRef{}, false
	}

	isSecretRef, ok := attrs[attrSecretRefIsSecretRef].(types.Bool)
	if !ok || isSecretRef.IsNull() || isSecretRef.IsUnknown() {
		return cloudConnectorSecretRef{}, false
	}

	return cloudConnectorSecretRef{
		ID:          id,
		IsSecretRef: isSecretRef,
	}, true
}
