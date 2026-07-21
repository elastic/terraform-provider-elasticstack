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

package managedintegration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustPackagePolicyFromJSON(t *testing.T, raw string) *kbapi.PackagePolicy {
	t.Helper()
	var pp kbapi.PackagePolicy
	require.NoError(t, json.Unmarshal([]byte(raw), &pp))
	return &pp
}

// baseTestModel returns a minimal agentlessPolicyModel with the identity and
// package attributes populated, matching what the entitycore envelope would
// have already decoded from plan/state before calling into conversion code.
func baseTestModel(t *testing.T) agentlessPolicyModel {
	t.Helper()
	ctx := context.Background()

	pkgObj, diags := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
		Name:    types.StringValue("cloud_security_posture"),
		Version: types.StringValue("3.4.0"),
		Title:   types.StringValue("Security Posture Management"),
	})
	require.False(t, diags.HasError())

	return agentlessPolicyModel{
		Name:                             types.StringValue("test-policy"),
		Package:                          pkgObj,
		Description:                      types.StringNull(),
		Namespace:                        types.StringNull(),
		PolicyTemplate:                   types.StringNull(),
		VarsJSON:                         policyshape.NewVarsJSONNull(),
		VarGroupSelections:               types.MapNull(types.StringType),
		Inputs:                           policyshape.NewInputsNull(agentlessInputType()),
		CloudConnector:                   types.ObjectNull(cloudConnectorAttrTypes()),
		GlobalDataTags:                   types.ListNull(types.ObjectType{AttrTypes: globalDataTagAttrTypes()}),
		AdditionalDatastreamsPermissions: types.ListNull(types.StringType),
		CreateDatasetTemplates:           types.BoolNull(),
		Force:                            types.BoolNull(),
		ForceDelete:                      types.BoolValue(false),
		SkipTopologyCheck:                types.BoolValue(false),
		SpaceIDs:                         types.SetNull(types.StringType),
	}
}

func decodeRequestJSON(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func TestToCreateBody_basicFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	m.PolicyID = types.StringValue("my-policy-id")
	m.Description = types.StringValue("a description")
	m.Namespace = types.StringValue("default")
	m.PolicyTemplate = types.StringValue("cspm")
	m.Force = types.BoolValue(true)
	m.CreateDatasetTemplates = types.BoolValue(true)

	body, diags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, diags.HasError(), "%v", diags)

	decoded := decodeRequestJSON(t, body)
	assert.Equal(t, "test-policy", decoded["name"])
	assert.Equal(t, "my-policy-id", decoded["id"])
	assert.Equal(t, "a description", decoded["description"])
	assert.Equal(t, "default", decoded["namespace"])
	assert.Equal(t, "cspm", decoded["policy_template"])
	assert.Equal(t, true, decoded["force"])
	assert.Equal(t, true, decoded["create_dataset_templates"])

	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cloud_security_posture", pkg["name"])
	assert.Equal(t, "3.4.0", pkg["version"])
	assert.Equal(t, "Security Posture Management", pkg["title"])
}

func TestToCreateBody_cloudConnectorOmittedWhenNotConfigured(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	// CloudConnector left as ObjectNull by baseTestModel -- block absent from config.

	body, diags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, diags.HasError(), "%v", diags)

	decoded := decodeRequestJSON(t, body)
	_, present := decoded["cloud_connector"]
	assert.False(t, present, "cloud_connector should be omitted entirely when the block is not present in config")
}

func TestToCreateBody_cloudConnectorSentWithEnabledFalse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	ccObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Enabled:          types.BoolValue(false),
		CloudConnectorID: types.StringNull(),
		Name:             types.StringNull(),
		TargetCSP:        types.StringNull(),
	})
	require.False(t, diags.HasError())
	m.CloudConnector = ccObj

	body, bodyDiags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	cc, ok := decoded["cloud_connector"].(map[string]any)
	require.True(t, ok, "cloud_connector should be sent (not elided) when the block is present in config")
	assert.Equal(t, false, cc["enabled"])
}

func TestToCreateBody_cloudConnectorWithID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	ccObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Enabled:          types.BoolValue(true),
		CloudConnectorID: types.StringValue("cc-abc123"),
		Name:             types.StringNull(),
		TargetCSP:        types.StringValue("aws"),
	})
	require.False(t, diags.HasError())
	m.CloudConnector = ccObj

	body, bodyDiags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	cc, ok := decoded["cloud_connector"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, cc["enabled"])
	assert.Equal(t, "cc-abc123", cc["cloud_connector_id"])
	assert.Equal(t, "aws", cc["target_csp"])
}

func TestToCreateBody_globalDataTags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	tagsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: globalDataTagAttrTypes()}, []globalDataTagModel{
		{Name: types.StringValue("env"), Value: types.StringValue("prod")},
	})
	require.False(t, diags.HasError())
	m.GlobalDataTags = tagsList

	body, bodyDiags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	tags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags, 1)
	tag, ok := tags[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "env", tag["name"])
	assert.Equal(t, "prod", tag["value"])
}

func TestToCreateBody_varsJSON(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	varsJSON, diags := policyshape.NewVarsJSONWithIntegration(`{"posture":"cspm","deployment":"aws"}`, "cloud_security_posture", "3.4.0", lookupCachedPackageInfo)
	require.False(t, diags.HasError())
	m.VarsJSON = varsJSON

	body, bodyDiags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cspm", vars["posture"])
	assert.Equal(t, "aws", vars["deployment"])
}

func TestToCreateBody_inputs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)

	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.account_type":"single-account"}`),
		},
	})
	require.False(t, diags.HasError())

	inputsValue, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"cloud_formation_template":"x"}`),
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	m.Inputs = inputsValue

	body, bodyDiags := m.toCreateBody(ctx, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	awsInput, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, awsInput["enabled"])

	vars, ok := awsInput["vars"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "x", vars["cloud_formation_template"])

	streams, ok := awsInput["streams"].(map[string]any)
	require.True(t, ok)
	findings, ok := streams["cloud_security_posture.findings"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, findings["enabled"])
	streamVars, ok := findings["vars"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "single-account", streamVars["aws.account_type"])
}

const mappedFormatPackagePolicyJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 1,
	"policy_id": "policy-1",
	"policy_ids": ["policy-1"],
	"spaceIds": ["default"],
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"vars": {"posture": "cspm", "deployment": "aws"},
	"description": "test description",
	"global_data_tags": [{"name": "env", "value": "prod"}],
	"additional_datastreams_permissions": ["logs-foo-*"],
	"var_group_selections": {"group1": "option1"},
	"inputs": {
		"cspm-cloudbeat/cis_aws": {
			"enabled": true,
			"vars": {"cloud_formation_template": "x"},
			"streams": {
				"cloud_security_posture.findings": {
					"enabled": true,
					"vars": {"aws.account_type": "single-account"}
				}
			}
		},
		"cspm-cloudbeat/cis_gcp": {
			"enabled": false
		}
	}
}`

func TestPopulateFromPackagePolicy_decodesMappedInputsAndFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	data := mustPackagePolicyFromJSON(t, mappedFormatPackagePolicyJSON)

	m := agentlessPolicyModel{
		Force:                  types.BoolValue(true),
		ForceDelete:            types.BoolValue(true),
		CreateDatasetTemplates: types.BoolValue(true),
		PolicyTemplate:         types.StringValue("cspm"),
		CloudConnector:         types.ObjectNull(cloudConnectorAttrTypes()),
	}

	diags := m.populateFromPackagePolicy(ctx, "default", data)
	require.False(t, diags.HasError(), "%v", diags)

	assert.Equal(t, "default/policy-1", m.ID.ValueString())
	assert.Equal(t, "policy-1", m.PolicyID.ValueString())
	assert.Equal(t, "test-policy", m.Name.ValueString())
	assert.Equal(t, "test description", m.Description.ValueString())
	assert.Equal(t, "default", m.Namespace.ValueString())
	assert.Equal(t, "2024-01-01T00:00:00.000Z", m.CreatedAt.ValueString())
	assert.Equal(t, "2024-01-02T00:00:00.000Z", m.UpdatedAt.ValueString())

	var spaceIDs []string
	require.False(t, m.SpaceIDs.ElementsAs(ctx, &spaceIDs, false).HasError())
	assert.Equal(t, []string{"default"}, spaceIDs)

	// Create-only / write-only flags must be preserved from the incoming
	// model, never sourced from the API response.
	assert.True(t, m.Force.ValueBool())
	assert.True(t, m.ForceDelete.ValueBool())
	assert.True(t, m.CreateDatasetTemplates.ValueBool())
	assert.Equal(t, "cspm", m.PolicyTemplate.ValueString())
	assert.True(t, m.CloudConnector.IsNull())

	var pkg packageModel
	require.False(t, m.Package.As(ctx, &pkg, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "cloud_security_posture", pkg.Name.ValueString())
	assert.Equal(t, "3.4.0", pkg.Version.ValueString())
	assert.Equal(t, "Security Posture Management", pkg.Title.ValueString())

	assert.Contains(t, m.VarsJSON.ValueString(), `"posture":"cspm"`)

	var vgs map[string]string
	require.False(t, m.VarGroupSelections.ElementsAs(ctx, &vgs, false).HasError())
	assert.Equal(t, "option1", vgs["group1"])

	var perms []string
	require.False(t, m.AdditionalDatastreamsPermissions.ElementsAs(ctx, &perms, false).HasError())
	assert.Equal(t, []string{"logs-foo-*"}, perms)

	var tags []globalDataTagModel
	require.False(t, m.GlobalDataTags.ElementsAs(ctx, &tags, false).HasError())
	require.Len(t, tags, 1)
	assert.Equal(t, "env", tags[0].Name.ValueString())
	assert.Equal(t, "prod", tags[0].Value.ValueString())

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	require.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	require.Contains(t, inputs, "cspm-cloudbeat/cis_gcp")
	assert.True(t, inputs["cspm-cloudbeat/cis_aws"].Enabled.ValueBool())
	assert.False(t, inputs["cspm-cloudbeat/cis_gcp"].Enabled.ValueBool())

	var streams map[string]policyshape.InputStreamModel
	require.False(t, inputs["cspm-cloudbeat/cis_aws"].Streams.ElementsAs(ctx, &streams, false).HasError())
	require.Contains(t, streams, "cloud_security_posture.findings")
	assert.Contains(t, streams["cloud_security_posture.findings"].Vars.ValueString(), "single-account")
}

// TestPopulateFromPackagePolicy_emptyDescriptionBecomesNull covers a real
// Kibana behavior (empirically confirmed against a live 9.4.3 deployment,
// see buildUpdateBody's comment): once a description is cleared via update,
// GET returns an explicit "" rather than omitting the field. description is
// Optional but not Computed in schema.go, so state must fold that "" back to
// null or a config that never set description would show a permanent diff.
func TestPopulateFromPackagePolicy_emptyDescriptionBecomesNull(t *testing.T) {
	t.Parallel()

	data := mustPackagePolicyFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"namespace": "",
		"enabled": true,
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"revision": 1,
		"description": "",
		"package": {"name": "cloud_security_posture", "version": "3.4.0"},
		"inputs": {}
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromPackagePolicy(context.Background(), "default", data)
	require.False(t, diags.HasError(), "%v", diags)

	assert.True(t, m.Description.IsNull(), "an explicit empty-string description from the API should fold to null")
	assert.True(t, m.Namespace.IsNull(), "an explicit empty-string namespace from the API should fold to null")
}

func TestPopulateFromPackagePolicy_nilData(t *testing.T) {
	t.Parallel()
	m := agentlessPolicyModel{Force: types.BoolValue(true)}
	diags := m.populateFromPackagePolicy(context.Background(), "default", nil)
	assert.False(t, diags.HasError())
	// Untouched -- caller (read.go) is responsible for treating a nil data
	// pointer as "not found" and not persisting state at all.
	assert.True(t, m.Force.ValueBool())
}

func TestPopulateFromCreateResponse_setsIdentityAndPreservesCreateOnlyFlags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := kbapi.KibanaHTTPAPIsManagedIntegration{
		Id:        "policy-2",
		Name:      "test-policy",
		CreatedAt: "2024-01-01T00:00:00.000Z",
		CreatedBy: "elastic",
		UpdatedAt: "2024-01-01T00:00:00.000Z",
		UpdatedBy: "elastic",
		Package: kbapi.KibanaHTTPAPIsManagedIntegrationPackage{
			Name:    "cloud_security_posture",
			Version: "3.4.0",
			Title:   "Security Posture Management",
		},
	}

	m := baseTestModel(t)
	m.Force = types.BoolValue(true)
	m.ForceDelete = types.BoolValue(true)
	m.CreateDatasetTemplates = types.BoolValue(true)
	m.PolicyTemplate = types.StringValue("cspm")

	diags := m.populateFromCreateResponse(ctx, "default", item)
	require.False(t, diags.HasError(), "%v", diags)

	assert.Equal(t, "default/policy-2", m.ID.ValueString())
	assert.Equal(t, "policy-2", m.PolicyID.ValueString())
	assert.Equal(t, "2024-01-01T00:00:00.000Z", m.CreatedAt.ValueString())

	var spaceIDs []string
	require.False(t, m.SpaceIDs.ElementsAs(ctx, &spaceIDs, false).HasError())
	assert.Equal(t, []string{"default"}, spaceIDs)

	assert.True(t, m.Force.ValueBool())
	assert.True(t, m.ForceDelete.ValueBool())
	assert.True(t, m.CreateDatasetTemplates.ValueBool())
	assert.Equal(t, "cspm", m.PolicyTemplate.ValueString())
}

// TestPopulateFromPackagePolicy_filtersToKnownInputKeys covers the
// inputsKnownKeySet/populateInputsModel fix for "Provider produced
// inconsistent result after apply": Fleet's package-policy responses for
// multi-policy-template packages like cloud_security_posture echo back every
// input the package declares across ALL of its policy templates, not just
// the one(s) actually configured (mappedFormatPackagePolicyJSON above
// includes both "cspm-cloudbeat/cis_aws" and "cspm-cloudbeat/cis_gcp" for
// exactly this reason). When the model's Inputs map is already Known with a
// specific key set (the plan's value on Update, or the prior state's value
// on Read), the decoded response must be filtered down to just those keys.
// TestPopulateFromPackagePolicy_decodesMappedInputsAndFields above starts
// m.Inputs at its Go zero value (an untyped, not-Known types.Map), so
// inputsKnownKeySet always took its nil/no-op path there; this test instead
// seeds a genuinely Known, single-key Inputs map before decoding a wire
// response containing two input keys, so the filtering path itself is
// exercised.
func TestPopulateFromPackagePolicy_filtersToKnownInputKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	data := mustPackagePolicyFromJSON(t, mappedFormatPackagePolicyJSON)

	knownInputs, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: types.MapNull(policyshape.StreamType()),
		},
	})
	require.False(t, diags.HasError())

	m := agentlessPolicyModel{
		CloudConnector: types.ObjectNull(cloudConnectorAttrTypes()),
		Inputs:         knownInputs,
	}

	popDiags := m.populateFromPackagePolicy(ctx, "default", data)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	assert.Len(t, inputs, 1, "only the previously-known input key should survive")
	assert.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.NotContains(t, inputs, "cspm-cloudbeat/cis_gcp",
		"cis_gcp is cross-policy-template noise from the wire response and was never in the known key set")
}

// TestPopulateFromCreateResponse_filtersToKnownInputKeys is the create-response
// counterpart of TestPopulateFromPackagePolicy_filtersToKnownInputKeys: the
// same cross-policy-template-noise behavior is present in the bundled
// POST /api/fleet/agentless_policies response (KibanaHTTPAPIsManagedIntegration),
// not just the package-policies GET/PUT response, and populateFromCreateResponse
// shares the same inputsKnownKeySet/populateInputsModel filtering path.
func TestPopulateFromCreateResponse_filtersToKnownInputKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := kbapi.KibanaHTTPAPIsManagedIntegration{
		Id:        "policy-2",
		Name:      "test-policy",
		CreatedAt: "2024-01-01T00:00:00.000Z",
		CreatedBy: "elastic",
		UpdatedAt: "2024-01-01T00:00:00.000Z",
		UpdatedBy: "elastic",
		Package: kbapi.KibanaHTTPAPIsManagedIntegrationPackage{
			Name:    "cloud_security_posture",
			Version: "3.4.0",
			Title:   "Security Posture Management",
		},
	}
	require.NoError(t, json.Unmarshal([]byte(`{
		"cspm-cloudbeat/cis_aws": {"enabled": true},
		"cspm-cloudbeat/cis_gcp": {"enabled": false},
		"kspm-cloudbeat/cis_k8s": {"enabled": false}
	}`), &item.Inputs))

	m := baseTestModel(t)
	knownInputs, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: types.MapNull(policyshape.StreamType()),
		},
	})
	require.False(t, diags.HasError())
	m.Inputs = knownInputs

	popDiags := m.populateFromCreateResponse(ctx, "default", item)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	assert.Len(t, inputs, 1, "only the previously-known input key should survive")
	assert.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.NotContains(t, inputs, "cspm-cloudbeat/cis_gcp",
		"cis_gcp is cross-policy-template noise from the wire response and was never in the known key set")
	assert.NotContains(t, inputs, "kspm-cloudbeat/cis_k8s",
		"kspm-cloudbeat/cis_k8s is cross-policy-template noise from the wire response and was never in the known key set")
}

func TestMappedInputKey(t *testing.T) {
	t.Parallel()

	cspm := "cspm"
	assert.Equal(t, "cspm-cloudbeat/cis_aws", mappedInputKey(&cspm, "cloudbeat/cis_aws"))

	empty := ""
	assert.Equal(t, "cloudbeat/cis_aws", mappedInputKey(&empty, "cloudbeat/cis_aws"))
	assert.Equal(t, "cloudbeat/cis_aws", mappedInputKey(nil, "cloudbeat/cis_aws"))
}

const typedFormatPackagePolicyJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 1,
	"policy_id": "policy-1",
	"policy_ids": ["policy-1"],
	"spaceIds": ["default"],
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"vars": {"posture": {"value": "cspm", "type": "text"}, "deployment": {"value": "aws", "type": "text"}},
	"description": "old description",
	"inputs": [
		{
			"type": "cloudbeat/cis_aws",
			"policy_template": "cspm",
			"enabled": true,
			"streams": [
				{
					"id": "cloudbeat/cis_aws-cloud_security_posture.findings-policy-1",
					"enabled": true,
					"data_stream": {"type": "logs", "dataset": "cloud_security_posture.findings"},
					"vars": {"aws.account_type": {"value": "single-account", "type": "text"}}
				}
			]
		},
		{
			"type": "cloudbeat/cis_gcp",
			"policy_template": "cspm",
			"enabled": false,
			"streams": [
				{
					"id": "cloudbeat/cis_gcp-cloud_security_posture.findings-policy-1",
					"enabled": false,
					"data_stream": {"type": "logs", "dataset": "cloud_security_posture.findings"}
				}
			]
		}
	]
}`

// typedFormatPackagePolicyMultiVarsJSON is typedFormatPackagePolicyJSON with
// a second per-input var (`input_var_a`/`input_var_b`) and a second
// per-stream var (`aws.account_type`/`aws.other_var`) added to the cis_aws
// input, so a plan that keeps only one of the two keys exercises *partial*
// removal (a strict-subset plan) rather than full removal (see
// TestBuildUpdateBody_partialVarsRemovalDropsOnlyMissingKeys). The top-level
// `vars` object already has two keys (`posture`/`deployment`) in the base
// fixture, so it is reused as-is.
const typedFormatPackagePolicyMultiVarsJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"enabled": true,
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"revision": 1,
	"policy_id": "policy-1",
	"policy_ids": ["policy-1"],
	"spaceIds": ["default"],
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"vars": {"posture": {"value": "cspm", "type": "text"}, "deployment": {"value": "aws", "type": "text"}},
	"description": "old description",
	"inputs": [
		{
			"type": "cloudbeat/cis_aws",
			"policy_template": "cspm",
			"enabled": true,
			"vars": {"input_var_a": {"value": "a", "type": "text"}, "input_var_b": {"value": "b", "type": "text"}},
			"streams": [
				{
					"id": "cloudbeat/cis_aws-cloud_security_posture.findings-policy-1",
					"enabled": true,
					"data_stream": {"type": "logs", "dataset": "cloud_security_posture.findings"},
					"vars": {
						"aws.account_type": {"value": "single-account", "type": "text"},
						"aws.other_var": {"value": "other", "type": "text"}
					}
				}
			]
		}
	]
}`

func TestBuildUpdateBody(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	plan := baseTestModel(t)
	plan.Description = types.StringValue("new description")

	varsJSON, diags := policyshape.NewVarsJSONWithIntegration(`{"posture":"cspm","deployment":"gcp"}`, "cloud_security_posture", "3.4.0", lookupCachedPackageInfo)
	require.False(t, diags.HasError())
	plan.VarsJSON = varsJSON

	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.account_type":"organization-account"}`),
		},
	})
	require.False(t, diags.HasError())

	inputsValue, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	plan.Inputs = inputsValue

	tagsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: globalDataTagAttrTypes()}, []globalDataTagModel{
		{Name: types.StringValue("env"), Value: types.StringValue("staging")},
	})
	require.False(t, diags.HasError())
	plan.GlobalDataTags = tagsList

	body, bodyDiags := buildUpdateBody(ctx, plan, current, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	assert.Equal(t, "test-policy", decoded["name"])
	assert.Equal(t, "default", decoded["namespace"])
	assert.Equal(t, "policy-1", decoded["policy_id"])
	assert.Equal(t, "new description", decoded["description"])

	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cloud_security_posture", pkg["name"])
	assert.Equal(t, "3.4.0", pkg["version"])
	// Here the plan's package object (built by baseTestModel) happens to
	// carry the same title as current's, so this assertion alone can't tell
	// a correct plan-overlay apart from a bug that always falls back to
	// current's title -- see TestBuildUpdateBody_packageTitleOverlay for a
	// case where the two differ.
	assert.Equal(t, "Security Posture Management", pkg["title"])

	tags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags, 1)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	deployment, ok := vars["deployment"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "gcp", deployment["value"])
	assert.Equal(t, "text", deployment["type"], "existing var `type` metadata should be preserved across the merge")
	posture, ok := vars["posture"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cspm", posture["value"])

	inputs, ok := decoded["inputs"].([]any)
	require.True(t, ok)
	require.Len(t, inputs, 2)

	var awsInput, gcpInput map[string]any
	for _, raw := range inputs {
		in, ok := raw.(map[string]any)
		require.True(t, ok)
		switch in["type"] {
		case "cloudbeat/cis_aws":
			awsInput = in
		case "cloudbeat/cis_gcp":
			gcpInput = in
		}
	}
	require.NotNil(t, awsInput)
	require.NotNil(t, gcpInput)

	// cis_gcp was not mentioned in the plan's inputs map: it is echoed back
	// unchanged (still disabled).
	assert.Equal(t, false, gcpInput["enabled"])

	awsStreams, ok := awsInput["streams"].([]any)
	require.True(t, ok)
	require.Len(t, awsStreams, 1)
	awsStream, ok := awsStreams[0].(map[string]any)
	require.True(t, ok)
	// data_stream metadata from the echoed GET response must survive the
	// round trip even though the plan doesn't model it.
	ds, ok := awsStream["data_stream"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cloud_security_posture.findings", ds["dataset"])

	streamVars, ok := awsStream["vars"].(map[string]any)
	require.True(t, ok)
	accountType, ok := streamVars["aws.account_type"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "organization-account", accountType["value"])
	assert.Equal(t, "text", accountType["type"], "existing stream var `type` metadata should be preserved across the merge")
}

// TestBuildUpdateBody_packageTitleOverlay closes the test gap left by
// TestBuildUpdateBody's package.title assertion: there, baseTestModel's plan
// title and current's fixture title happen to be identical
// ("Security Posture Management"), so that assertion alone can't distinguish
// a correct plan-overlay from a bug that always falls back to current's
// title. Here the two differ, so the PUT body must carry the plan's value.
func TestBuildUpdateBody_packageTitleOverlay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	plan := baseTestModel(t)
	pkgObj, diags := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
		Name:    types.StringValue("cloud_security_posture"),
		Version: types.StringValue("3.4.0"),
		Title:   types.StringValue("Custom CSPM Title"),
	})
	require.False(t, diags.HasError())
	plan.Package = pkgObj

	body, bodyDiags := buildUpdateBody(ctx, plan, current, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Custom CSPM Title", pkg["title"], "package.title must come from the plan, not fall back to current's title")
}

// TestBuildUpdateBody_clearsVarsWhenPlanRemovesThem covers a bug found in
// review: `vars` (top-level, per-input, and per-stream) is Optional but not
// Computed in schema.go, so removing a `vars` block from config must clear
// the corresponding API value entirely, not silently echo back whatever the
// fetched typed snapshot already had.
func TestBuildUpdateBody_clearsVarsWhenPlanRemovesThem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyJSON)

	plan := baseTestModel(t)
	plan.Description = types.StringValue("old description")
	// VarsJSON left as NewVarsJSONNull() by baseTestModel: an explicit
	// `vars_json = null` in config must clear current's top-level vars.

	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			// Vars left as the jsontypes.Normalized zero value (null): the
			// plan removed this stream's `vars` block.
		},
	})
	require.False(t, diags.HasError())

	inputsValue, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			// Vars left null too: the plan removed the input's own `vars` block.
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	plan.Inputs = inputsValue

	body, bodyDiags := buildUpdateBody(ctx, plan, current, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, vars, "vars_json = null should clear the top-level vars, not echo current's")

	inputs, ok := decoded["inputs"].([]any)
	require.True(t, ok)

	var awsInput map[string]any
	for _, raw := range inputs {
		in, ok := raw.(map[string]any)
		require.True(t, ok)
		if in["type"] == "cloudbeat/cis_aws" {
			awsInput = in
		}
	}
	require.NotNil(t, awsInput)
	assert.Empty(t, awsInput["vars"], "removing an input's vars block should clear it, not echo current's")

	awsStreams, ok := awsInput["streams"].([]any)
	require.True(t, ok)
	require.Len(t, awsStreams, 1)
	awsStream, ok := awsStreams[0].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, awsStream["vars"], "removing a stream's vars block should clear it, not echo current's")
}

// TestBuildUpdateBody_partialVarsRemovalDropsOnlyMissingKeys covers the
// [BLOCKER] regression found in review: mergeVarsInto used to seed its
// result from *dst (the vars already on the policy) and only overlay keys
// present in the plan, so a key present in existing vars but absent from a
// non-empty plan survived forever -- users could never *reduce* the key set
// of vars_json / an input's vars / a stream's vars via Update, only add to
// or overwrite it. This is distinct from
// TestBuildUpdateBody_clearsVarsWhenPlanRemovesThem, which covers *full*
// removal (the whole vars block absent from the plan); here the plan sets a
// non-empty vars object that is a strict subset of the existing keys, at all
// three levels mergeVarsInto is used: top-level vars_json, per-input vars,
// and per-stream vars.
func TestBuildUpdateBody_partialVarsRemovalDropsOnlyMissingKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	current := mustPackagePolicyFromJSON(t, typedFormatPackagePolicyMultiVarsJSON)

	plan := baseTestModel(t)
	plan.Description = types.StringValue("old description")

	// Top-level: current has posture+deployment; plan keeps only posture.
	varsJSON, diags := policyshape.NewVarsJSONWithIntegration(`{"posture":"cspm"}`, "cloud_security_posture", "3.4.0", lookupCachedPackageInfo)
	require.False(t, diags.HasError())
	plan.VarsJSON = varsJSON

	// Stream-level: current has aws.account_type+aws.other_var; plan keeps
	// only aws.account_type.
	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.account_type":"single-account"}`),
		},
	})
	require.False(t, diags.HasError())

	// Input-level: current has input_var_a+input_var_b; plan keeps only
	// input_var_a.
	inputsValue, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"input_var_a":"a"}`),
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	plan.Inputs = inputsValue

	body, bodyDiags := buildUpdateBody(ctx, plan, current, agentlessPolicyFeatures{SupportsCondition: true})
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, vars, 1, "top-level vars must contain exactly the plan's keys, not existing's")
	assert.Contains(t, vars, "posture")
	assert.NotContains(t, vars, "deployment", "deployment was dropped from the plan and must not survive the merge")
	posture, ok := vars["posture"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "text", posture["type"], "surviving key's existing `type` metadata should be preserved")

	inputs, ok := decoded["inputs"].([]any)
	require.True(t, ok)

	var awsInput map[string]any
	for _, raw := range inputs {
		in, ok := raw.(map[string]any)
		require.True(t, ok)
		if in["type"] == "cloudbeat/cis_aws" {
			awsInput = in
		}
	}
	require.NotNil(t, awsInput)

	inputVars, ok := awsInput["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, inputVars, 1, "input vars must contain exactly the plan's keys, not existing's")
	assert.Contains(t, inputVars, "input_var_a")
	assert.NotContains(t, inputVars, "input_var_b", "input_var_b was dropped from the plan and must not survive the merge")

	awsStreams, ok := awsInput["streams"].([]any)
	require.True(t, ok)
	require.Len(t, awsStreams, 1)
	awsStream, ok := awsStreams[0].(map[string]any)
	require.True(t, ok)

	streamVars, ok := awsStream["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, streamVars, 1, "stream vars must contain exactly the plan's keys, not existing's")
	assert.Contains(t, streamVars, "aws.account_type")
	assert.NotContains(t, streamVars, "aws.other_var", "aws.other_var was dropped from the plan and must not survive the merge")
	accountType, ok := streamVars["aws.account_type"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "text", accountType["type"], "surviving key's existing `type` metadata should be preserved")
}
