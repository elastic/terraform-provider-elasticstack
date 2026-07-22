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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustManagedIntegrationFromJSON(t *testing.T, raw string) *kbapi.KibanaHTTPAPIsManagedIntegration {
	t.Helper()
	var item kbapi.KibanaHTTPAPIsManagedIntegration
	require.NoError(t, json.Unmarshal([]byte(raw), &item))
	return &item
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
		GlobalDataTags:                   types.MapNull(globalDataTagsElementType()),
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

	body, diags := m.toCreateBody(ctx)
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

	body, diags := m.toCreateBody(ctx)
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

	body, bodyDiags := m.toCreateBody(ctx)
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

	body, bodyDiags := m.toCreateBody(ctx)
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
	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"env": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringValue("prod"),
			globalDataTagNumberValueAttr: types.Float32Null(),
		}),
	})
	require.False(t, diags.HasError())
	m.GlobalDataTags = tagsMap

	body, bodyDiags := m.toCreateBody(ctx)
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

func TestToCreateBody_globalDataTags_number(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"priority": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringNull(),
			globalDataTagNumberValueAttr: types.Float32Value(42),
		}),
	})
	require.False(t, diags.HasError())
	m.GlobalDataTags = tagsMap

	body, bodyDiags := m.toCreateBody(ctx)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	tags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags, 1)
	tag, ok := tags[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "priority", tag["name"])
	assert.InDelta(t, float64(42), tag["value"], 0.001)
}

// TestGlobalDataTags_mapAttribute_mixedStringAndNumberRoundTrip exercises the
// Terraform map shape (keys = tag names, values = {string_value | number_value})
// through API encode (globalDataTagsRawFromModel) and decode
// (globalDataTagsToModel via populateFromManagedIntegration).
func TestGlobalDataTags_mapAttribute_mixedStringAndNumberRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"env": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringValue("prod"),
			globalDataTagNumberValueAttr: types.Float32Null(),
		}),
		"priority": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringNull(),
			globalDataTagNumberValueAttr: types.Float32Value(7.5),
		}),
	})
	require.False(t, diags.HasError())

	m := baseTestModel(t)
	m.GlobalDataTags = tagsMap

	body, bodyDiags := m.toCreateBody(ctx)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)
	decoded := decodeRequestJSON(t, body)
	apiTags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, apiTags, 2)
	tagByName := make(map[string]map[string]any, len(apiTags))
	for _, rawTag := range apiTags {
		tag, ok := rawTag.(map[string]any)
		require.True(t, ok)
		name, ok := tag["name"].(string)
		require.True(t, ok)
		tagByName[name] = tag
	}
	require.Contains(t, tagByName, "env")
	require.Contains(t, tagByName, "priority")
	assert.Equal(t, "prod", tagByName["env"]["value"])
	assert.InDelta(t, float64(7.5), tagByName["priority"]["value"], 0.001)

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"global_data_tags": [
			{"name": "env", "value": "prod"},
			{"name": "priority", "value": 7.5}
		]
	}`)

	out := agentlessPolicyModel{}
	popDiags := out.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var tags map[string]globalDataTagsItemModel
	require.False(t, out.GlobalDataTags.ElementsAs(ctx, &tags, false).HasError())
	require.Len(t, tags, 2)
	assert.Equal(t, "prod", tags["env"].StringValue.ValueString())
	assert.True(t, tags["env"].NumberValue.IsNull())
	assert.InDelta(t, float32(7.5), tags["priority"].NumberValue.ValueFloat32(), 0.001)
	assert.True(t, tags["priority"].StringValue.IsNull())

	var encodeDiags diag.Diagnostics
	raw := globalDataTagsRawFromModel(ctx, out.GlobalDataTags, &encodeDiags)
	require.False(t, encodeDiags.HasError(), "%v", encodeDiags)
	require.NotNil(t, raw)
	require.Len(t, *raw, 2)
	encodedByName := make(map[string]struct {
		stringVal string
		numberVal float32
		hasString bool
		hasNumber bool
	}, 2)
	for _, entry := range *raw {
		switch entry.Name {
		case "env":
			s, err := entry.Value.AsKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue0()
			require.NoError(t, err)
			encodedByName["env"] = struct {
				stringVal string
				numberVal float32
				hasString bool
				hasNumber bool
			}{stringVal: s, hasString: true}
		case "priority":
			n, err := entry.Value.AsKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue1()
			require.NoError(t, err)
			encodedByName["priority"] = struct {
				stringVal string
				numberVal float32
				hasString bool
				hasNumber bool
			}{numberVal: n, hasNumber: true}
		default:
			t.Fatalf("unexpected global_data_tags name %q in encode output", entry.Name)
		}
	}
	require.True(t, encodedByName["env"].hasString)
	assert.Equal(t, "prod", encodedByName["env"].stringVal)
	require.True(t, encodedByName["priority"].hasNumber)
	assert.InDelta(t, float32(7.5), encodedByName["priority"].numberVal, 0.001)
}

func TestGlobalDataTagsToModel_numberWire(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "x",
		"name": "n",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "p", "version": "1.0.0", "title": "t"},
		"global_data_tags": [{"name": "priority", "value": 1.5}]
	}`)

	var diags diag.Diagnostics
	m := globalDataTagsToModel(ctx, item, &diags)
	require.False(t, diags.HasError(), "%v", diags)

	var tags map[string]globalDataTagsItemModel
	require.False(t, m.ElementsAs(ctx, &tags, false).HasError())
	require.InDelta(t, float32(1.5), tags["priority"].NumberValue.ValueFloat32(), 0.001)
	assert.True(t, tags["priority"].StringValue.IsNull())
}

func TestGlobalDataTagsToModel_unsupportedWireType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "x",
		"name": "n",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "p", "version": "1.0.0", "title": "t"},
		"global_data_tags": [{"name": "bad", "value": ["not", "a", "scalar"]}]
	}`)

	var diags diag.Diagnostics
	m := globalDataTagsToModel(ctx, item, &diags)
	assert.True(t, diags.HasError())
	assert.True(t, m.IsNull())
}

func TestGlobalDataTagsRawFromModel_number(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"priority": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringNull(),
			globalDataTagNumberValueAttr: types.Float32Value(7),
		}),
	})
	require.False(t, diags.HasError())

	var encodeDiags diag.Diagnostics
	raw := globalDataTagsRawFromModel(ctx, tagsMap, &encodeDiags)
	require.False(t, encodeDiags.HasError(), "%v", encodeDiags)
	require.NotNil(t, raw)
	require.Len(t, *raw, 1)
	assert.Equal(t, "priority", (*raw)[0].Name)
	num, err := (*raw)[0].Value.AsKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue1()
	require.NoError(t, err)
	assert.InDelta(t, float32(7), num, 0.001)
}

func TestGlobalDataTagsRawFromModel_neitherValueSet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"my_tag": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringNull(),
			globalDataTagNumberValueAttr: types.Float32Null(),
		}),
	})
	require.False(t, diags.HasError())

	var encodeDiags diag.Diagnostics
	raw := globalDataTagsRawFromModel(ctx, tagsMap, &encodeDiags)
	assert.True(t, encodeDiags.HasError())
	assert.Nil(t, raw)
}

func TestToCreateBody_varsJSON(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	varsJSON, diags := policyshape.NewVarsJSONWithIntegration(`{"posture":"cspm","deployment":"aws"}`, "cloud_security_posture", "3.4.0", lookupCachedPackageInfo)
	require.False(t, diags.HasError())
	m.VarsJSON = varsJSON

	body, bodyDiags := m.toCreateBody(ctx)
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

	body, bodyDiags := m.toCreateBody(ctx)
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

const mappedFormatManagedIntegrationJSON = `{
	"id": "policy-1",
	"name": "test-policy",
	"namespace": "default",
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
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

func TestPopulateFromManagedIntegration_decodesInputsAndFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, mappedFormatManagedIntegrationJSON)

	m := agentlessPolicyModel{
		Force:                  types.BoolValue(true),
		ForceDelete:            types.BoolValue(true),
		CreateDatasetTemplates: types.BoolValue(true),
		PolicyTemplate:         types.StringValue("cspm"),
		CloudConnector:         types.ObjectNull(cloudConnectorAttrTypes()),
	}

	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
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

	var tags map[string]globalDataTagsItemModel
	require.False(t, m.GlobalDataTags.ElementsAs(ctx, &tags, false).HasError())
	require.Len(t, tags, 1)
	assert.Equal(t, "prod", tags["env"].StringValue.ValueString())

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

// TestPopulateFromManagedIntegration_emptyDescriptionBecomesNull covers a real
// Kibana behavior: once a description is cleared via update, GET returns an
// explicit "" rather than omitting the field.
func TestPopulateFromManagedIntegration_emptyDescriptionBecomesNull(t *testing.T) {
	t.Parallel()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"namespace": "",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"description": "",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"inputs": {}
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromManagedIntegration(context.Background(), "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	assert.True(t, m.Description.IsNull(), "an explicit empty-string description from the API should fold to null")
	assert.True(t, m.Namespace.IsNull(), "an explicit empty-string namespace from the API should fold to null")
}

func TestPopulateFromManagedIntegration_nilData(t *testing.T) {
	t.Parallel()
	m := agentlessPolicyModel{Force: types.BoolValue(true)}
	diags := m.populateFromManagedIntegration(context.Background(), "default", nil, nil)
	assert.False(t, diags.HasError())
	assert.True(t, m.Force.ValueBool())
}

func TestPopulateFromManagedIntegration_globalDataTagsNumberRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"global_data_tags": [{"name": "priority", "value": 42}]
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var tags map[string]globalDataTagsItemModel
	require.False(t, m.GlobalDataTags.ElementsAs(ctx, &tags, false).HasError())
	require.InDelta(t, float32(42), tags["priority"].NumberValue.ValueFloat32(), 0.001)

	encodeDiags := diag.Diagnostics{}
	raw := globalDataTagsRawFromModel(ctx, m.GlobalDataTags, &encodeDiags)
	require.False(t, encodeDiags.HasError(), "%v", encodeDiags)
	require.NotNil(t, raw)
	require.Len(t, *raw, 1)
	num, err := (*raw)[0].Value.AsKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue1()
	require.NoError(t, err)
	assert.InDelta(t, float32(42), num, 0.001)
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

	diags := m.populateFromManagedIntegration(ctx, "default", &item, nil)
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

// TestPopulateFromManagedIntegration_filtersToKnownInputKeys covers the
// inputsKnownKeySet filtering behavior for multi-policy-template packages.
func TestPopulateFromManagedIntegration_filtersToKnownInputKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, mappedFormatManagedIntegrationJSON)

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

	popDiags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var inputs map[string]agentlessInputModel
	require.False(t, m.Inputs.ElementsAs(ctx, &inputs, false).HasError())
	assert.Len(t, inputs, 1, "only the previously-known input key should survive")
	assert.Contains(t, inputs, "cspm-cloudbeat/cis_aws")
	assert.NotContains(t, inputs, "cspm-cloudbeat/cis_gcp",
		"cis_gcp is cross-policy-template noise from the wire response and was never in the known key set")
}

// TestPopulateFromCreateResponse_filtersToKnownInputKeys is the create-response
// counterpart of TestPopulateFromManagedIntegration_filtersToKnownInputKeys.
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

	popDiags := m.populateFromManagedIntegration(ctx, "default", &item, nil)
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

func TestToCreateBody_omitsInputsWhenNullOrUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	m := baseTestModel(t)
	body, diags := m.toCreateBody(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	decoded := decodeRequestJSON(t, body)
	_, present := decoded["inputs"]
	assert.False(t, present, "null inputs in config must omit the inputs field from the create body")

	m.Inputs = policyshape.InputsValue{MapValue: types.MapUnknown(agentlessInputType())}
	body, diags = m.toCreateBody(ctx)
	require.False(t, diags.HasError(), "%v", diags)
	decoded = decodeRequestJSON(t, body)
	_, present = decoded["inputs"]
	assert.False(t, present, "unknown inputs in config must omit the inputs field from the create body")
}

func TestPopulateFromManagedIntegration_emptyInputsNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"inputs": {}
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, m.Inputs.IsNull())
}

func TestPopulateFromManagedIntegration_omittedGlobalDataTagsNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"}
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)
	assert.True(t, m.GlobalDataTags.IsNull())
}

func TestPopulateFromManagedIntegration_stringGlobalDataTagsReadEncodeRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"global_data_tags": [{"name": "env", "value": "prod"}]
	}`)

	m := agentlessPolicyModel{}
	diags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, diags.HasError(), "%v", diags)

	var tags map[string]globalDataTagsItemModel
	require.False(t, m.GlobalDataTags.ElementsAs(ctx, &tags, false).HasError())
	assert.Equal(t, "prod", tags["env"].StringValue.ValueString())

	encodeDiags := diag.Diagnostics{}
	raw := globalDataTagsRawFromModel(ctx, m.GlobalDataTags, &encodeDiags)
	require.False(t, encodeDiags.HasError(), "%v", encodeDiags)
	require.NotNil(t, raw)
	require.Len(t, *raw, 1)
	str, err := (*raw)[0].Value.AsKibanaHTTPAPIsCreateManagedIntegrationRequestGlobalDataTagsValue0()
	require.NoError(t, err)
	assert.Equal(t, "prod", str)
}

func TestPopulateFromManagedIntegration_preservesCloudConnector(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ccObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Enabled:          types.BoolValue(true),
		CloudConnectorID: types.StringValue("cc-123"),
		Name:             types.StringValue("my-connector"),
		TargetCSP:        types.StringValue("aws"),
	})
	require.False(t, diags.HasError())

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"cloud_connector": {"enabled": true, "cloud_connector_id": "cc-other"}
	}`)

	m := agentlessPolicyModel{CloudConnector: ccObj}
	popDiags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, popDiags.HasError(), "%v", popDiags)

	var cc cloudConnectorModel
	require.False(t, m.CloudConnector.As(ctx, &cc, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "cc-other", cc.CloudConnectorID.ValueString())
	assert.Equal(t, "my-connector", cc.Name.ValueString())
	assert.Equal(t, "aws", cc.TargetCSP.ValueString())
}

func TestPopulateFromManagedIntegration_explicitSpaceIDs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"}
	}`)

	m := agentlessPolicyModel{}
	spaceIDs := []string{"space-a", "space-b"}
	diags := m.populateFromManagedIntegration(ctx, "default", item, &spaceIDs)
	require.False(t, diags.HasError(), "%v", diags)

	var ids []string
	require.False(t, m.SpaceIDs.ElementsAs(ctx, &ids, false).HasError())
	assert.ElementsMatch(t, []string{"space-a", "space-b"}, ids)
}

func TestGlobalDataTagsToModel_duplicateNames(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	item := mustManagedIntegrationFromJSON(t, `{
		"id": "x",
		"name": "n",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-01T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "p", "version": "1.0.0", "title": "t"},
		"global_data_tags": [
			{"name": "env", "value": "prod"},
			{"name": "env", "value": "staging"}
		]
	}`)

	var diags diag.Diagnostics
	m := globalDataTagsToModel(ctx, item, &diags)
	assert.True(t, diags.HasError())
	assert.True(t, m.IsNull())
	requireDiagnosticAtPath(t, diags, path.Root("global_data_tags").AtMapKey("env"), "Duplicate global_data_tags name")
}

func TestMappedInputKey(t *testing.T) {
	t.Parallel()

	cspm := "cspm"
	assert.Equal(t, "cspm-cloudbeat/cis_aws", mappedInputKey(&cspm, "cloudbeat/cis_aws"))

	empty := ""
	assert.Equal(t, "cloudbeat/cis_aws", mappedInputKey(&empty, "cloudbeat/cis_aws"))
	assert.Equal(t, "cloudbeat/cis_aws", mappedInputKey(nil, "cloudbeat/cis_aws"))
}

func TestBuildUpdateBody(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	prior.Namespace = types.StringValue("default")

	plan := prior
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

	tagsMap, diags := types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"env": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringValue("staging"),
			globalDataTagNumberValueAttr: types.Float32Null(),
		}),
	})
	require.False(t, diags.HasError())
	plan.GlobalDataTags = tagsMap

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	assert.Equal(t, "test-policy", decoded["name"])
	assert.Equal(t, "default", decoded["namespace"])
	_, hasPolicyID := decoded["id"]
	assert.False(t, hasPolicyID, "update body must not re-send create-only id")
	assert.Equal(t, "new description", decoded["description"])

	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cloud_security_posture", pkg["name"])
	assert.Equal(t, "3.4.0", pkg["version"])
	assert.Equal(t, "Security Posture Management", pkg["title"])

	tags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags, 1)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "gcp", vars["deployment"])
	assert.Equal(t, "cspm", vars["posture"])

	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	require.Len(t, inputs, 1)

	awsInput, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, awsInput["enabled"])

	streams, ok := awsInput["streams"].(map[string]any)
	require.True(t, ok)
	awsStream, ok := streams["cloud_security_posture.findings"].(map[string]any)
	require.True(t, ok)
	streamVars, ok := awsStream["vars"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "organization-account", streamVars["aws.account_type"])
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

	prior := baseTestModel(t)
	plan := prior
	pkgObj, diags := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
		Name:    types.StringValue("cloud_security_posture"),
		Version: types.StringValue("3.4.0"),
		Title:   types.StringValue("Custom CSPM Title"),
	})
	require.False(t, diags.HasError())
	plan.Package = pkgObj

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Custom CSPM Title", pkg["title"], "package.title must come from the plan")
}

// TestBuildUpdateBody_clearsVarsWhenPlanRemovesThem covers a bug found in
// review: `vars` (top-level, per-input, and per-stream) is Optional but not
// Computed in schema.go, so removing a `vars` block from config must clear
// the corresponding API value entirely, not silently echo back whatever the
// fetched typed snapshot already had.
func TestBuildUpdateBody_clearsVarsWhenPlanRemovesThem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	plan := prior
	plan.Description = types.StringValue("old description")
	// VarsJSON left as NewVarsJSONNull() by baseTestModel: an explicit
	// `vars_json = null` in config must clear top-level vars on full replace.

	streamsMap, diags := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
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

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, vars, "vars_json = null should clear the top-level vars")

	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	awsInput, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, awsInput["vars"], "removing an input's vars block should clear it")

	streams, ok := awsInput["streams"].(map[string]any)
	require.True(t, ok)
	awsStream, ok := streams["cloud_security_posture.findings"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, awsStream["vars"], "removing a stream's vars block should clear it")
}

// TestBuildUpdateBody_partialVarsRemovalDropsOnlyMissingKeys covers full-replace
// semantics: the PUT body vars maps contain exactly the plan's keys at the
// top level, per-input, and per-stream.
func TestBuildUpdateBody_partialVarsRemovalDropsOnlyMissingKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	plan := prior
	plan.Description = types.StringValue("old description")

	varsJSON, diags := policyshape.NewVarsJSONWithIntegration(`{"posture":"cspm"}`, "cloud_security_posture", "3.4.0", lookupCachedPackageInfo)
	require.False(t, diags.HasError())
	plan.VarsJSON = varsJSON

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
			Vars:    jsontypes.NewNormalizedValue(`{"input_var_a":"a"}`),
			Streams: streamsMap,
		},
	})
	require.False(t, diags.HasError())
	plan.Inputs = inputsValue

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)

	vars, ok := decoded["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, vars, 1)
	assert.Contains(t, vars, "posture")
	assert.NotContains(t, vars, "deployment")

	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	awsInput, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)

	inputVars, ok := awsInput["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, inputVars, 1)
	assert.Contains(t, inputVars, "input_var_a")
	assert.NotContains(t, inputVars, "input_var_b")

	streams, ok := awsInput["streams"].(map[string]any)
	require.True(t, ok)
	awsStream, ok := streams["cloud_security_posture.findings"].(map[string]any)
	require.True(t, ok)

	streamVars, ok := awsStream["vars"].(map[string]any)
	require.True(t, ok)
	assert.Len(t, streamVars, 1)
	assert.Contains(t, streamVars, "aws.account_type")
	assert.NotContains(t, streamVars, "aws.other_var")
}
