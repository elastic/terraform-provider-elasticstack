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

// Package cloudconnector provides Terraform Plugin Framework models and API
// conversion helpers for the elasticstack_fleet_cloud_connector resource.
package cloudconnector

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	attrID                    = "id"
	attrCloudConnectorID      = "cloud_connector_id"
	attrSpaceID               = "space_id"
	attrName                  = "name"
	attrCloudProvider         = "cloud_provider"
	attrAccountType           = "account_type"
	attrForceDelete           = "force_delete"
	attrAWSBlock              = "aws"
	attrAzureBlock            = "azure"
	attrVarsMap               = "vars"
	attrNamespace             = "namespace"
	attrPackagePolicyCount    = "package_policy_count"
	attrVerificationStatus    = "verification_status"
	attrVerificationStartedAt = "verification_started_at"
	attrVerificationFailedAt  = "verification_failed_at"
	attrCreatedAt             = "created_at"
	attrUpdatedAt             = "updated_at"

	attrVarsString      = "string"
	attrVarsNumber      = "number"
	attrVarsBool        = "bool"
	attrVarsType        = "type"
	attrVarsFrozen      = "frozen"
	attrVarsValue       = "value"
	attrVarsSecretValue = "secret_value"
	attrVarsSecretRef   = "secret_ref"

	attrSecretRefID          = "id"
	attrSecretRefIsSecretRef = "is_secret_ref"

	attrAWSRoleArn             = "role_arn"
	attrAWSExternalID          = "external_id"
	attrAWSExternalIDSecretRef = "external_id_secret_ref"

	attrAzureTenantID         = "tenant_id"
	attrAzureClientID         = "client_id"
	attrAzureCloudConnectorID = "cloud_connector_id"

	cloudProviderAWS   = attrAWSBlock
	cloudProviderAzure = attrAzureBlock

	wireKeyFrozen      = "frozen"
	wireKeyValue       = "value"
	wireKeyIsSecretRef = "isSecretRef"
)

type cloudConnectorModel struct {
	ID                    types.String `tfsdk:"id"`
	KibanaConnection      types.List   `tfsdk:"kibana_connection"`
	CloudConnectorID      types.String `tfsdk:"cloud_connector_id"`
	SpaceID               types.String `tfsdk:"space_id"`
	Name                  types.String `tfsdk:"name"`
	CloudProvider         types.String `tfsdk:"cloud_provider"`
	AccountType           types.String `tfsdk:"account_type"`
	ForceDelete           types.Bool   `tfsdk:"force_delete"`
	AWS                   types.Object `tfsdk:"aws"`
	Azure                 types.Object `tfsdk:"azure"`
	Vars                  types.Map    `tfsdk:"vars"`
	Namespace             types.String `tfsdk:"namespace"`
	PackagePolicyCount    types.Int64  `tfsdk:"package_policy_count"`
	VerificationStatus    types.String `tfsdk:"verification_status"`
	VerificationStartedAt types.String `tfsdk:"verification_started_at"`
	VerificationFailedAt  types.String `tfsdk:"verification_failed_at"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

// cloudConnectorVarsElement models one entry in the vars map union. Task 5 will
// define the matching ObjectType schema from varsElementAttrTypes().
//
// secret_value is write-only and never populated from API reads. Write-only
// drift detection uses bcrypt hashes in private state (Decision 5); there is
// no plan-visible secret_value_wo_version companion on this model.
type cloudConnectorVarsElement struct {
	String types.String  `tfsdk:"string"`
	Number types.Float64 `tfsdk:"number"` // wire is float32; Float64 for PF schema compatibility
	Bool   types.Bool    `tfsdk:"bool"`

	Type   types.String `tfsdk:"type"`
	Frozen types.Bool   `tfsdk:"frozen"`

	Value       types.String `tfsdk:"value"`
	SecretValue types.String `tfsdk:"secret_value"`
	SecretRef   types.Object `tfsdk:"secret_ref"`
}

type cloudConnectorSecretRef struct {
	ID          types.String `tfsdk:"id"`
	IsSecretRef types.Bool   `tfsdk:"is_secret_ref"`
}

type awsBlockModel struct {
	RoleArn             types.String `tfsdk:"role_arn"`
	ExternalID          types.String `tfsdk:"external_id"`
	ExternalIDSecretRef types.Object `tfsdk:"external_id_secret_ref"`
}

type azureBlockModel struct {
	TenantID         types.String `tfsdk:"tenant_id"`
	ClientID         types.String `tfsdk:"client_id"`
	CloudConnectorID types.String `tfsdk:"cloud_connector_id"`
}

var cloudConnectorMinVersion = version.Must(version.NewVersion("9.2.0"))

func (m cloudConnectorModel) GetID() types.String             { return m.ID }
func (m cloudConnectorModel) GetResourceID() types.String     { return m.CloudConnectorID }
func (m cloudConnectorModel) GetSpaceID() types.String        { return m.SpaceID }
func (m cloudConnectorModel) GetKibanaConnection() types.List { return m.KibanaConnection }

func (m cloudConnectorModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *cloudConnectorMinVersion,
			ErrorMessage: fmt.Sprintf("Fleet cloud connectors require Kibana v%s or later.", cloudConnectorMinVersion),
		},
	}, nil
}

func varsElementAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrVarsString:      types.StringType,
		attrVarsNumber:      types.Float64Type,
		attrVarsBool:        types.BoolType,
		attrVarsType:        types.StringType,
		attrVarsFrozen:      types.BoolType,
		attrVarsValue:       types.StringType,
		attrVarsSecretValue: types.StringType,
		attrVarsSecretRef:   types.ObjectType{AttrTypes: secretRefAttrTypes()},
	}
}

func secretRefAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrSecretRefID:          types.StringType,
		attrSecretRefIsSecretRef: types.BoolType,
	}
}

func awsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAWSRoleArn:             types.StringType,
		attrAWSExternalID:          types.StringType,
		attrAWSExternalIDSecretRef: types.ObjectType{AttrTypes: secretRefAttrTypes()},
	}
}

func azureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrAzureTenantID:         types.StringType,
		attrAzureClientID:         types.StringType,
		attrAzureCloudConnectorID: types.StringType,
	}
}

func (m *cloudConnectorModel) populateFromAPI(spaceID string, item fleetclient.CloudConnectorItem) diag.Diagnostics {
	// Does not modify config-only attributes such as ForceDelete.
	var diags diag.Diagnostics

	m.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: item.ID}).String())
	m.CloudConnectorID = types.StringValue(item.ID)
	m.SpaceID = types.StringValue(spaceID)
	m.Name = types.StringValue(item.Name)
	m.CloudProvider = types.StringValue(item.CloudProvider)

	if item.AccountType != nil && *item.AccountType != "" {
		m.AccountType = types.StringValue(*item.AccountType)
	} else {
		m.AccountType = types.StringNull()
	}

	if item.Namespace != nil && *item.Namespace != "" {
		m.Namespace = types.StringValue(*item.Namespace)
	} else {
		m.Namespace = types.StringNull()
	}

	m.PackagePolicyCount = types.Int64Value(int64(item.PackagePolicyCount))

	if item.VerificationStatus != nil && *item.VerificationStatus != "" {
		m.VerificationStatus = types.StringValue(*item.VerificationStatus)
	} else {
		m.VerificationStatus = types.StringNull()
	}

	if item.VerificationStartedAt != nil && *item.VerificationStartedAt != "" {
		m.VerificationStartedAt = types.StringValue(*item.VerificationStartedAt)
	} else {
		m.VerificationStartedAt = types.StringNull()
	}

	if item.VerificationFailedAt != nil && *item.VerificationFailedAt != "" {
		m.VerificationFailedAt = types.StringValue(*item.VerificationFailedAt)
	} else {
		m.VerificationFailedAt = types.StringNull()
	}

	if item.CreatedAt != "" {
		m.CreatedAt = types.StringValue(item.CreatedAt)
	} else {
		m.CreatedAt = types.StringNull()
	}

	if item.UpdatedAt != "" {
		m.UpdatedAt = types.StringValue(item.UpdatedAt)
	} else {
		m.UpdatedAt = types.StringNull()
	}

	varsMap, varsDiags := varsMapToModel(item.Vars)
	diags.Append(varsDiags...)
	if diags.HasError() {
		return diags
	}
	m.Vars = varsMap

	switch item.CloudProvider {
	case cloudProviderAWS:
		awsObj, awsDiags := awsBlockFromVars(item.Vars)
		diags.Append(awsDiags...)
		if diags.HasError() {
			return diags
		}
		m.AWS = awsObj
		m.Azure = types.ObjectNull(azureAttrTypes())
	case cloudProviderAzure:
		azureObj, azureDiags := azureBlockFromVars(item.Vars)
		diags.Append(azureDiags...)
		if diags.HasError() {
			return diags
		}
		m.Azure = azureObj
		m.AWS = types.ObjectNull(awsAttrTypes())
	default:
		m.AWS = types.ObjectNull(awsAttrTypes())
		m.Azure = types.ObjectNull(azureAttrTypes())
	}

	return diags
}

func varsMapToModel(vars map[string]any) (types.Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(vars) == 0 {
		empty, mapDiags := types.MapValue(types.ObjectType{AttrTypes: varsElementAttrTypes()}, map[string]attr.Value{})
		diags.Append(mapDiags...)
		return empty, diags
	}

	elems := make(map[string]attr.Value, len(vars))
	for key, value := range vars {
		elem, elemDiags := varValueToElement(key, value)
		diags.Append(elemDiags...)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
		}

		obj, objDiags := varsElementToObject(elem)
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}), diags
		}
		elems[key] = obj
	}

	varsMap, mapDiags := types.MapValue(types.ObjectType{AttrTypes: varsElementAttrTypes()}, elems)
	diags.Append(mapDiags...)
	return varsMap, diags
}

func varValueToElement(key string, value any) (cloudConnectorVarsElement, diag.Diagnostics) {
	var diags diag.Diagnostics
	nullElem := cloudConnectorVarsElement{
		String:      types.StringNull(),
		Number:      types.Float64Null(),
		Bool:        types.BoolNull(),
		Type:        types.StringNull(),
		Frozen:      types.BoolNull(),
		Value:       types.StringNull(),
		SecretValue: types.StringNull(),
		SecretRef:   types.ObjectNull(secretRefAttrTypes()),
	}

	if value == nil {
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q is null; the provider cannot represent null var values in state.", key),
		)
		return nullElem, diags
	}

	switch v := value.(type) {
	case string:
		elem := nullElem
		elem.String = types.StringValue(v)
		return elem, diags
	case bool:
		elem := nullElem
		elem.Bool = types.BoolValue(v)
		return elem, diags
	case float64:
		elem := nullElem
		elem.Number = types.Float64Value(v)
		return elem, diags
	case float32:
		elem := nullElem
		elem.Number = types.Float64Value(float64(v))
		return elem, diags
	case int:
		elem := nullElem
		elem.Number = types.Float64Value(float64(v))
		return elem, diags
	case int64:
		elem := nullElem
		elem.Number = types.Float64Value(float64(v))
		return elem, diags
	case map[string]any:
		if _, hasType := v[attrVarsType]; hasType {
			return structuredVarToElement(key, v)
		}
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q has an object value without a type field; the provider cannot represent it in state.", key),
		)
		return nullElem, diags
	default:
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q has value type %T; the provider cannot represent it in state.", key, value),
		)
		return nullElem, diags
	}
}

func structuredVarToElement(key string, structured map[string]any) (cloudConnectorVarsElement, diag.Diagnostics) {
	var diags diag.Diagnostics
	nullElem := cloudConnectorVarsElement{
		String:      types.StringNull(),
		Number:      types.Float64Null(),
		Bool:        types.BoolNull(),
		Type:        types.StringNull(),
		Frozen:      types.BoolNull(),
		Value:       types.StringNull(),
		SecretValue: types.StringNull(),
		SecretRef:   types.ObjectNull(secretRefAttrTypes()),
	}

	typeVal, ok := structured[attrVarsType].(string)
	if !ok || typeVal == "" {
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q structured value is missing a string type field.", key),
		)
		return nullElem, diags
	}

	elem := nullElem
	elem.Type = types.StringValue(typeVal)

	if frozenVal, ok := structured[wireKeyFrozen]; ok && frozenVal != nil {
		switch f := frozenVal.(type) {
		case bool:
			elem.Frozen = types.BoolValue(f)
		default:
			diags.AddError(
				"Unsupported cloud connector var value",
				fmt.Sprintf("Cloud connector var %q frozen field has type %T; expected bool.", key, frozenVal),
			)
			return nullElem, diags
		}
	}

	valueVal, hasValue := structured[wireKeyValue]
	if !hasValue {
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q structured value is missing a value field.", key),
		)
		return nullElem, diags
	}

	switch vv := valueVal.(type) {
	case string:
		elem.Value = types.StringValue(vv)
	case map[string]any:
		ref, refDiags := secretRefFromMap(key, vv)
		diags.Append(refDiags...)
		if diags.HasError() {
			return nullElem, diags
		}
		elem.SecretRef = ref
	default:
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q structured value field has type %T; expected string or secret reference object.", key, valueVal),
		)
		return nullElem, diags
	}

	return elem, diags
}

func secretRefFromMap(key string, ref map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	idVal, ok := ref[attrSecretRefID].(string)
	if !ok || idVal == "" {
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q secret reference is missing a string id field.", key),
		)
		return types.ObjectNull(secretRefAttrTypes()), diags
	}

	isSecretRefVal, ok := ref[wireKeyIsSecretRef].(bool)
	if !ok {
		diags.AddError(
			"Unsupported cloud connector var value",
			fmt.Sprintf("Cloud connector var %q secret reference is missing a bool isSecretRef field.", key),
		)
		return types.ObjectNull(secretRefAttrTypes()), diags
	}

	obj, objDiags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue(idVal),
		IsSecretRef: types.BoolValue(isSecretRefVal),
	})
	diags.Append(objDiags...)
	return obj, diags
}

func secretRefToObject(ref cloudConnectorSecretRef) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(secretRefAttrTypes(), map[string]attr.Value{
		attrSecretRefID:          ref.ID,
		attrSecretRefIsSecretRef: ref.IsSecretRef,
	})
}

func varsElementToObject(elem cloudConnectorVarsElement) (types.Object, diag.Diagnostics) {
	return types.ObjectValue(varsElementAttrTypes(), map[string]attr.Value{
		attrVarsString:      elem.String,
		attrVarsNumber:      elem.Number,
		attrVarsBool:        elem.Bool,
		attrVarsType:        elem.Type,
		attrVarsFrozen:      elem.Frozen,
		attrVarsValue:       elem.Value,
		attrVarsSecretValue: elem.SecretValue,
		attrVarsSecretRef:   elem.SecretRef,
	})
}

func awsBlockFromVars(vars map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	nullObj := types.ObjectNull(awsAttrTypes())

	if !hasExactVarKeys(vars, attrAWSRoleArn, attrAWSExternalID) {
		return nullObj, diags
	}

	roleArnVal := vars[attrAWSRoleArn]
	externalIDVal := vars[attrAWSExternalID]

	roleArn, ok := stringVarValue(roleArnVal)
	if !ok {
		diags.AddError(
			"Unsupported cloud connector AWS block",
			`Cloud connector var "role_arn" is not a string or structured string value.`,
		)
		return nullObj, diags
	}

	block := awsBlockModel{
		RoleArn:             types.StringValue(roleArn),
		ExternalID:          types.StringNull(),
		ExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	}

	if secretRef, ok := secretRefVarValue(externalIDVal); ok {
		block.ExternalIDSecretRef = secretRef
	}

	obj, objDiags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             block.RoleArn,
		attrAWSExternalID:          block.ExternalID,
		attrAWSExternalIDSecretRef: block.ExternalIDSecretRef,
	})
	diags.Append(objDiags...)
	return obj, diags
}

func azureBlockFromVars(vars map[string]any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	nullObj := types.ObjectNull(azureAttrTypes())

	if !hasExactVarKeys(vars, attrAzureTenantID, attrAzureClientID, attrAzureCloudConnectorID) {
		return nullObj, diags
	}

	tenantIDVal := vars[attrAzureTenantID]
	clientIDVal := vars[attrAzureClientID]
	connectorIDVal := vars[attrAzureCloudConnectorID]

	tenantID, ok := stringVarValue(tenantIDVal)
	if !ok {
		diags.AddError(
			"Unsupported cloud connector Azure block",
			`Cloud connector var "tenant_id" is not a string or structured string value.`,
		)
		return nullObj, diags
	}

	clientID, ok := stringVarValue(clientIDVal)
	if !ok {
		diags.AddError(
			"Unsupported cloud connector Azure block",
			`Cloud connector var "client_id" is not a string or structured string value.`,
		)
		return nullObj, diags
	}

	cloudConnectorID, ok := stringVarValue(connectorIDVal)
	if !ok {
		diags.AddError(
			"Unsupported cloud connector Azure block",
			`Cloud connector var "cloud_connector_id" is not a string or structured string value.`,
		)
		return nullObj, diags
	}

	block := azureBlockModel{
		TenantID:         types.StringValue(tenantID),
		ClientID:         types.StringValue(clientID),
		CloudConnectorID: types.StringValue(cloudConnectorID),
	}

	obj, objDiags := types.ObjectValue(azureAttrTypes(), map[string]attr.Value{
		attrAzureTenantID:         block.TenantID,
		attrAzureClientID:         block.ClientID,
		attrAzureCloudConnectorID: block.CloudConnectorID,
	})
	diags.Append(objDiags...)
	return obj, diags
}

func stringVarValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case map[string]any:
		if _, hasType := v[attrVarsType]; !hasType {
			return "", false
		}
		valueVal, ok := v[wireKeyValue]
		if !ok {
			return "", false
		}
		s, ok := valueVal.(string)
		return s, ok
	default:
		return "", false
	}
}

func secretRefVarValue(value any) (types.Object, bool) {
	structured, ok := value.(map[string]any)
	if !ok {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	if _, hasType := structured[attrVarsType]; !hasType {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	refMap, ok := structured[wireKeyValue].(map[string]any)
	if !ok {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	idVal, ok := refMap[attrSecretRefID].(string)
	if !ok || idVal == "" {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	isSecretRefVal, ok := refMap[wireKeyIsSecretRef].(bool)
	if !ok {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	obj, diags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue(idVal),
		IsSecretRef: types.BoolValue(isSecretRefVal),
	})
	if diags.HasError() {
		return types.ObjectNull(secretRefAttrTypes()), false
	}
	return obj, true
}

// hasExactVarKeys reports whether vars is non-nil, contains every required key,
// and contains no keys outside the required set. A nil or empty map is exact
// only when no keys are required.
func hasExactVarKeys(vars map[string]any, requiredKeys ...string) bool {
	if len(vars) != len(requiredKeys) {
		return false
	}
	for _, key := range requiredKeys {
		if _, ok := vars[key]; !ok {
			return false
		}
	}
	return true
}

var (
	_ entitycore.KibanaResourceModel     = cloudConnectorModel{}
	_ entitycore.WithVersionRequirements = cloudConnectorModel{}
)
