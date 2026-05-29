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
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type modifyPlanFixture struct {
	t      *testing.T
	r      *Resource
	schema rschema.Schema
}

func newModifyPlanFixture(t *testing.T) *modifyPlanFixture {
	t.Helper()

	ctx := context.Background()
	r := newResource()
	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	require.False(t, sr.Diagnostics.HasError())

	return &modifyPlanFixture{
		t:      t,
		r:      r,
		schema: sr.Schema,
	}
}

func completeAWSCloudConnectorModel(t *testing.T, externalID types.String) cloudConnectorModel {
	t.Helper()

	return cloudConnectorModel{
		ID:               types.StringValue("default/connector-1"),
		KibanaConnection: providerschema.KibanaConnectionNullList(),
		CloudConnectorID: types.StringValue("connector-1"),
		SpaceID:          types.StringValue("default"),
		Name:             types.StringValue("test-connector"),
		CloudProvider:    types.StringValue("aws"),
		ForceDelete:      types.BoolValue(false),
		AWS:              mustAWSBlockObject(t, externalID),
		Azure:            types.ObjectNull(azureAttrTypes()),
		Vars:             types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}),
		UpdatedAt:        types.StringValue("2024-01-01T00:00:00Z"),
	}
}

func completeVarsCloudConnectorModel(t *testing.T, vars types.Map) cloudConnectorModel {
	t.Helper()

	return cloudConnectorModel{
		ID:               types.StringValue("default/connector-1"),
		KibanaConnection: providerschema.KibanaConnectionNullList(),
		CloudConnectorID: types.StringValue("connector-1"),
		SpaceID:          types.StringValue("default"),
		Name:             types.StringValue("test-connector"),
		CloudProvider:    types.StringValue("gcp"),
		ForceDelete:      types.BoolValue(false),
		AWS:              types.ObjectNull(awsAttrTypes()),
		Azure:            types.ObjectNull(azureAttrTypes()),
		Vars:             vars,
		UpdatedAt:        types.StringValue("2024-01-01T00:00:00Z"),
	}
}

func (f *modifyPlanFixture) modelToConfig(model cloudConnectorModel) tfsdk.Config {
	f.t.Helper()
	return tfsdk.Config(f.modelToPlan(model))
}

func (f *modifyPlanFixture) modelToState(model cloudConnectorModel) tfsdk.State {
	f.t.Helper()

	ctx := context.Background()
	st := tfsdk.State{Schema: f.schema}
	diags := st.Set(ctx, &model)
	require.False(f.t, diags.HasError(), diags)
	return st
}

func (f *modifyPlanFixture) modelToPlan(model cloudConnectorModel) tfsdk.Plan {
	f.t.Helper()

	ctx := context.Background()
	pl := tfsdk.Plan{Schema: f.schema}
	diags := pl.Set(ctx, &model)
	require.False(f.t, diags.HasError(), diags)
	return pl
}

func (f *modifyPlanFixture) nullRootValue() tftypes.Value {
	f.t.Helper()
	return tftypes.NewValue(f.schema.Type().TerraformType(context.Background()), nil)
}

func (f *modifyPlanFixture) run(
	config cloudConnectorModel,
	state tfsdk.State,
	plan tfsdk.Plan,
	priv *mapPrivateState,
) *resource.ModifyPlanResponse {
	f.t.Helper()

	if priv != nil {
		f.r.testModifyPlanPrivate = priv
	}

	resp := &resource.ModifyPlanResponse{Plan: plan}
	ctx := context.Background()
	f.r.ModifyPlan(ctx, resource.ModifyPlanRequest{
		Config: f.modelToConfig(config),
		State:  state,
		Plan:   plan,
	}, resp)

	f.r.testModifyPlanPrivate = nil
	return resp
}

func planUpdatedAt(ctx context.Context, t *testing.T, plan tfsdk.Plan) types.String {
	t.Helper()

	var updatedAt types.String
	diags := plan.GetAttribute(ctx, path.Root(attrUpdatedAt), &updatedAt)
	require.False(t, diags.HasError(), diags)
	return updatedAt
}

func TestResource_ModifyPlan(t *testing.T) {
	t.Parallel()

	hasher := cloudConnectorHasher()
	const (
		secretA = "secret-alpha"
		secretB = "secret-beta"
	)

	t.Run("destroy plan early return", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		model := completeAWSCloudConnectorModel(t, types.StringValue(secretA))
		state := f.modelToState(model)
		plan := tfsdk.Plan{Schema: f.schema, Raw: f.nullRootValue()}

		resp := f.run(model, state, plan, newMapPrivateState())
		require.False(t, resp.Diagnostics.HasError())
		assert.Empty(t, resp.Diagnostics.Warnings())
	})

	t.Run("create plan early return", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		model := completeAWSCloudConnectorModel(t, types.StringValue(secretA))
		plan := f.modelToPlan(model)
		state := tfsdk.State{Schema: f.schema, Raw: f.nullRootValue()}

		resp := f.run(model, state, plan, newMapPrivateState())
		require.False(t, resp.Diagnostics.HasError())
		assert.Empty(t, resp.Diagnostics.Warnings())
	})

	t.Run("no drift preserves updated_at", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		priv := newMapPrivateState()
		hash, err := hasher.Compute(secretA)
		require.NoError(t, err)
		priv.data[awsExternalIDPrivateStateKey()] = hash

		model := completeAWSCloudConnectorModel(t, types.StringValue(secretA))
		resp := f.run(model, f.modelToState(model), f.modelToPlan(model), priv)
		require.False(t, resp.Diagnostics.HasError())
		assert.Empty(t, resp.Diagnostics.Warnings())

		updatedAt := planUpdatedAt(context.Background(), t, resp.Plan)
		assert.Equal(t, "2024-01-01T00:00:00Z", updatedAt.ValueString())
	})

	t.Run("aws external_id changed", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		priv := newMapPrivateState()
		hash, err := hasher.Compute(secretA)
		require.NoError(t, err)
		priv.data[awsExternalIDPrivateStateKey()] = hash

		stateModel := completeAWSCloudConnectorModel(t, types.StringNull())
		configModel := completeAWSCloudConnectorModel(t, types.StringValue(secretB))
		resp := f.run(configModel, f.modelToState(stateModel), f.modelToPlan(configModel), priv)
		require.False(t, resp.Diagnostics.HasError())
		require.NotEmpty(t, resp.Diagnostics.Warnings())

		warn := resp.Diagnostics.Warnings()[0]
		assert.Contains(t, warn.Summary(), writeOnlyAttributeAWSExternalID)
		assert.Contains(t, warn.Summary(), "Detected a change")
		assert.NotContains(t, warn.Summary(), secretB)
		assert.NotContains(t, warn.Detail(), secretB)

		updatedAt := planUpdatedAt(context.Background(), t, resp.Plan)
		assert.True(t, updatedAt.IsUnknown())
	})

	t.Run("import baseline", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		model := completeAWSCloudConnectorModel(t, types.StringValue(secretA))
		resp := f.run(model, f.modelToState(model), f.modelToPlan(model), newMapPrivateState())
		require.False(t, resp.Diagnostics.HasError())
		require.NotEmpty(t, resp.Diagnostics.Warnings())

		warn := resp.Diagnostics.Warnings()[0]
		assert.Contains(t, warn.Summary(), writeOnlyAttributeAWSExternalID)
		assert.Contains(t, warn.Summary(), "Establishing baseline hash")
		assert.NotContains(t, warn.Summary(), secretA)
		assert.NotContains(t, warn.Detail(), secretA)

		updatedAt := planUpdatedAt(context.Background(), t, resp.Plan)
		assert.True(t, updatedAt.IsUnknown())
	})

	t.Run("vars secret_value changed", func(t *testing.T) {
		t.Parallel()

		f := newModifyPlanFixture(t)
		priv := newMapPrivateState()
		hash, err := hasher.Compute("old-var-secret")
		require.NoError(t, err)
		priv.data[varsSecretValuePrivateStateKey("token")] = hash
		priv.data[varsSecretIndexPrivateStateKey] = []byte(`["token"]`)

		stateModel := completeVarsCloudConnectorModel(t, mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"token": {
				Type:        types.StringValue("password"),
				SecretValue: types.StringNull(),
			},
		}))
		configModel := completeVarsCloudConnectorModel(t, mustVarsMap(t, map[string]cloudConnectorVarsElement{
			"token": {
				Type:        types.StringValue("password"),
				SecretValue: types.StringValue("new-var-secret"),
			},
		}))
		resp := f.run(configModel, f.modelToState(stateModel), f.modelToPlan(configModel), priv)
		require.False(t, resp.Diagnostics.HasError())
		require.NotEmpty(t, resp.Diagnostics.Warnings())

		attrPath := varsSecretValueAttributePath("token")
		warn := resp.Diagnostics.Warnings()[0]
		assert.Contains(t, warn.Summary(), attrPath)
		assert.Contains(t, warn.Summary(), "Detected a change")
		assert.NotContains(t, warn.Summary(), "new-var-secret")
		assert.NotContains(t, warn.Detail(), "new-var-secret")

		updatedAt := planUpdatedAt(context.Background(), t, resp.Plan)
		assert.True(t, updatedAt.IsUnknown())
	})
}

func mustAzureBlockObject(t *testing.T) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(azureAttrTypes(), map[string]attr.Value{
		attrAzureTenantID:         types.StringValue("tenant-1"),
		attrAzureClientID:         types.StringValue("client-1"),
		attrAzureCloudConnectorID: types.StringValue("azure-conn-1"),
	})
	require.False(t, diags.HasError())
	return obj
}

func mustAWSRoleArnVarsMap(t *testing.T, roleArn string) types.Map {
	t.Helper()
	return mustVarsMap(t, map[string]cloudConnectorVarsElement{
		attrAWSRoleArn: {
			Type:  types.StringValue("text"),
			Value: types.StringValue(roleArn),
		},
		attrAWSExternalID: {
			Type:      types.StringValue("password"),
			SecretRef: mustSecretRefObject(t, "secret-ref-1"),
		},
	})
}

func mustAzureVarsMap(t *testing.T) types.Map {
	t.Helper()
	return mustVarsMap(t, map[string]cloudConnectorVarsElement{
		attrAzureTenantID: {
			Type:  types.StringValue("text"),
			Value: types.StringValue("tenant-1"),
		},
		attrAzureClientID: {
			Type:  types.StringValue("text"),
			Value: types.StringValue("client-1"),
		},
		attrAzureCloudConnectorID: {
			Type:  types.StringValue("text"),
			Value: types.StringValue("azure-conn-1"),
		},
	})
}

func mustSecretRefObject(t *testing.T, id string) types.Object {
	t.Helper()
	obj, diags := secretRefToObject(cloudConnectorSecretRef{
		ID:          types.StringValue(id),
		IsSecretRef: types.BoolValue(true),
	})
	require.False(t, diags.HasError())
	return obj
}

func mustAWSBlockObjectWithRoleArn(t *testing.T, roleArn string) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue(roleArn),
		attrAWSExternalID:          types.StringNull(),
		attrAWSExternalIDSecretRef: mustSecretRefObject(t, "secret-ref-1"),
	})
	require.False(t, diags.HasError())
	return obj
}

func planAttributeObject(ctx context.Context, t *testing.T, plan tfsdk.Plan, attrName string) types.Object {
	t.Helper()
	var obj types.Object
	diags := plan.GetAttribute(ctx, path.Root(attrName), &obj)
	require.False(t, diags.HasError(), diags)
	return obj
}

func planAttributeMap(ctx context.Context, t *testing.T, plan tfsdk.Plan, attrName string) types.Map {
	t.Helper()
	var m types.Map
	diags := plan.GetAttribute(ctx, path.Root(attrName), &m)
	require.False(t, diags.HasError(), diags)
	return m
}

func TestResource_ModifyPlan_DualPopulationPreservesVarsWhenAWSConfigured(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	ctx := context.Background()

	config := completeAWSCloudConnectorModel(t, types.StringValue("secret"))
	config.Vars = types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()})

	state := config
	state.Vars = mustAWSRoleArnVarsMap(t, "arn:aws:iam::123:role/x")

	plan := config
	plan.Vars = types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()})

	resp := f.run(config, f.modelToState(state), f.modelToPlan(plan), newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())

	planVars := planAttributeMap(ctx, t, resp.Plan, attrVarsMap)
	assert.True(t, planVars.Equal(state.Vars))
}

func TestResource_ModifyPlan_DualPopulationPreservesAWSWhenVarsConfigured(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	ctx := context.Background()

	vars := mustAWSRoleArnVarsMap(t, "arn:aws:iam::123:role/x")
	config := completeVarsCloudConnectorModel(t, vars)
	config.CloudProvider = types.StringValue("aws")

	state := config
	state.AWS = mustAWSBlockObjectWithRoleArn(t, "arn:aws:iam::123:role/x")

	plan := config
	plan.AWS = types.ObjectNull(awsAttrTypes())

	resp := f.run(config, f.modelToState(state), f.modelToPlan(plan), newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())

	planAWS := planAttributeObject(ctx, t, resp.Plan, attrAWSBlock)
	assert.True(t, planAWS.Equal(state.AWS))
}

func TestResource_ModifyPlan_DualPopulationPreservesVarsWhenAzureConfigured(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	ctx := context.Background()

	config := cloudConnectorModel{
		ID:               types.StringValue("default/connector-1"),
		KibanaConnection: providerschema.KibanaConnectionNullList(),
		CloudConnectorID: types.StringValue("connector-1"),
		SpaceID:          types.StringValue("default"),
		Name:             types.StringValue("test-connector"),
		CloudProvider:    types.StringValue("azure"),
		ForceDelete:      types.BoolValue(false),
		Azure:            mustAzureBlockObject(t),
		AWS:              types.ObjectNull(awsAttrTypes()),
		Vars:             types.MapNull(types.ObjectType{AttrTypes: varsElementAttrTypes()}),
		UpdatedAt:        types.StringValue("2024-01-01T00:00:00Z"),
	}

	state := config
	state.Vars = mustAzureVarsMap(t)

	plan := config

	resp := f.run(config, f.modelToState(state), f.modelToPlan(plan), newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())

	planVars := planAttributeMap(ctx, t, resp.Plan, attrVarsMap)
	assert.True(t, planVars.Equal(state.Vars))
}

func TestResource_ModifyPlan_DualPopulationPreservesAzureWhenVarsConfigured(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	ctx := context.Background()

	vars := mustAzureVarsMap(t)
	config := completeVarsCloudConnectorModel(t, vars)
	config.CloudProvider = types.StringValue("azure")

	state := config
	state.Azure = mustAzureBlockObject(t)

	plan := config
	plan.Azure = types.ObjectNull(azureAttrTypes())

	resp := f.run(config, f.modelToState(state), f.modelToPlan(plan), newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())

	planAzure := planAttributeObject(ctx, t, resp.Plan, attrAzureBlock)
	assert.True(t, planAzure.Equal(state.Azure))
}

func TestResource_ModifyPlan_DualPopulationOnCreateSkipsReconciliation(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	model := completeAWSCloudConnectorModel(t, types.StringValue("secret"))
	plan := f.modelToPlan(model)
	state := tfsdk.State{Schema: f.schema, Raw: f.nullRootValue()}

	resp := f.run(model, state, plan, newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())
}

func TestResource_ModifyPlan_DualPopulationOnDestroySkipsReconciliation(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	model := completeAWSCloudConnectorModel(t, types.StringValue("secret"))
	stateModel := model
	stateModel.Vars = mustAWSRoleArnVarsMap(t, "arn:aws:iam::123:role/x")
	state := f.modelToState(stateModel)
	plan := tfsdk.Plan{Schema: f.schema, Raw: f.nullRootValue()}

	resp := f.run(model, state, plan, newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())
}

func TestResource_ModifyPlan_UserChangesConfigDoesNotRestoreOldSibling(t *testing.T) {
	t.Parallel()

	f := newModifyPlanFixture(t)
	ctx := context.Background()

	stateVars := mustAWSRoleArnVarsMap(t, "arn:aws:iam::123:role/old")
	configVars := mustAWSRoleArnVarsMap(t, "arn:aws:iam::123:role/new")

	config := completeVarsCloudConnectorModel(t, configVars)
	config.CloudProvider = types.StringValue("aws")

	state := completeVarsCloudConnectorModel(t, stateVars)
	state.CloudProvider = types.StringValue("aws")
	state.AWS = mustAWSBlockObjectWithRoleArn(t, "arn:aws:iam::123:role/old")

	plan := config
	plan.AWS = types.ObjectNull(awsAttrTypes())

	resp := f.run(config, f.modelToState(state), f.modelToPlan(plan), newMapPrivateState())
	require.False(t, resp.Diagnostics.HasError())

	planAWS := planAttributeObject(ctx, t, resp.Plan, attrAWSBlock)
	assert.True(t, planAWS.IsNull())
}
