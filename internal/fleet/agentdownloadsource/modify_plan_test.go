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

package agentdownloadsource

import (
	"context"
	"testing"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestPlanDriftWarnings(t *testing.T) {
	t.Parallel()

	hash, err := secretHasher.Compute("applied")
	require.NoError(t, err)
	stored := map[string][]byte{sslKeyWOPath: hash}

	tests := []struct {
		name   string
		config map[string]types.String
		stored map[string][]byte
		want   []string
	}{
		{name: "unchanged value", config: map[string]types.String{sslKeyWOPath: types.StringValue("applied")}, stored: stored},
		{
			name:   "changed value",
			config: map[string]types.String{sslKeyWOPath: types.StringValue("rotated")},
			stored: stored,
			want:   []string{sslKeyWOPath},
		},
		{
			name:   "removed value with stored hash",
			config: map[string]types.String{sslKeyWOPath: types.StringNull()},
			stored: stored,
			want:   []string{sslKeyWOPath},
		},
		{name: "removed value without stored hash", config: map[string]types.String{sslKeyWOPath: types.StringNull()}},
		{
			name:   "configured without stored hash (re-added or added to an existing resource)",
			config: map[string]types.String{sslKeyWOPath: types.StringValue("new")},
			want:   []string{sslKeyWOPath},
		},
		{
			name:   "unknown value (upstream-generated rotation) schedules an update",
			config: map[string]types.String{sslKeyWOPath: types.StringUnknown()},
			stored: stored,
			want:   []string{sslKeyWOPath},
		},
		{
			name:   "absent attribute with stored hash counts as removed",
			config: map[string]types.String{},
			stored: stored,
			want:   []string{sslKeyWOPath},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, planDriftPaths(tt.config, tt.stored))
		})
	}
}

func fullSchema(ctx context.Context, t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	newResource().Schema(ctx, resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())
	return resp
}

// fullModel returns a model that serializes against the full resource schema.
func fullModel(ctx context.Context, t *testing.T, m model) model {
	t.Helper()
	s := fullSchema(ctx, t).Schema
	m.Timeouts = timeouts.Value{Object: types.ObjectNull(s.Attributes["timeouts"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes())}
	m.KibanaConnection = providerschema.KibanaConnectionNullList()
	return m
}

func tfsdkNullRaw(ctx context.Context, s schema.Schema) tftypes.Value {
	return tftypes.NewValue(s.Type().TerraformType(ctx), nil)
}

func planAndConfig(ctx context.Context, t *testing.T, planModel, configModel model) (tfsdk.Plan, tfsdk.Config) {
	t.Helper()
	s := fullSchema(ctx, t).Schema
	planModel, configModel = fullModel(ctx, t, planModel), fullModel(ctx, t, configModel)

	empty := tfsdk.State{Schema: s, Raw: tfsdkNullRaw(ctx, s)}
	require.False(t, empty.Set(ctx, &planModel).HasError())
	plan := tfsdk.Plan{Schema: s, Raw: empty.Raw}

	cfgState := tfsdk.State{Schema: s, Raw: tfsdkNullRaw(ctx, s)}
	require.False(t, cfgState.Set(ctx, &configModel).HasError())
	return plan, tfsdk.Config{Schema: s, Raw: cfgState.Raw}
}

func baseModel(ctx context.Context, keyWO types.String) model {
	return model{
		ID:       types.StringValue("src"),
		SourceID: types.StringValue("src"),
		Name:     types.StringValue("n"),
		Host:     types.StringValue("https://h"),
		Default:  types.BoolValue(false),
		SpaceIDs: types.SetNull(types.StringType),
		Ssl:      testSslObject(ctx, types.StringNull(), keyWO),
	}
}

func TestModifyPlanForcesUpdateOnDriftWithoutDeletingHash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	private := newFakePrivate()
	var seed diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{sslKeyWOPath: "applied"}, nil, &seed)
	before := private.data[secretHashKey(sslKeyWOPath)]

	tests := []struct {
		name       string
		configKey  types.String
		wantUpdate bool
	}{
		{name: "unchanged", configKey: types.StringValue("applied")},
		{name: "changed", configKey: types.StringValue("rotated"), wantUpdate: true},
		{name: "removed", configKey: types.StringNull(), wantUpdate: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, config := planAndConfig(ctx, t, baseModel(ctx, types.StringNull()), baseModel(ctx, tt.configKey))
			diags := modifyPlan(ctx, private, config, &plan)
			require.False(t, diags.HasError(), "unexpected diagnostics: %#v", diags)

			var id types.String
			require.False(t, plan.GetAttribute(ctx, path.Root("id"), &id).HasError())
			require.Equal(t, tt.wantUpdate, id.IsUnknown())
			require.Equal(t, tt.wantUpdate, len(diags) > 0, "a detected change is surfaced as a warning")
			require.Equal(t, before, private.data[secretHashKey(sslKeyWOPath)], "ModifyPlan must not touch the stored hash")
		})
	}
}

func TestModifyPlanConfiguredValueWithoutStoredHashPlansUpdate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	plan, config := planAndConfig(ctx, t, baseModel(ctx, types.StringNull()), baseModel(ctx, types.StringValue("first")))
	diags := modifyPlan(ctx, newFakePrivate(), config, &plan)
	require.False(t, diags.HasError())
	require.Len(t, diags, 1, "the change is surfaced as a warning")

	var id types.String
	require.False(t, plan.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.True(t, id.IsUnknown(), "a write-only value that was never applied must force an update")
}

func TestModifyPlanReadsAuthWriteOnlyAttributesFromConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	private := newFakePrivate()
	var seed diag.Diagnostics
	persistSecretHashes(ctx, private, map[string]string{authPasswordWOPath: "pw", authAPIKeyWOPath: "ak"}, nil, &seed)

	build := func(passwordWO, apiKeyWO types.String) model {
		m := baseModel(ctx, types.StringNull())
		m.Ssl = types.ObjectNull(getSslAttrTypes(ctx))
		m.Auth = &authModel{
			Username: types.StringValue("u"), Password: types.StringNull(), PasswordWO: passwordWO,
			APIKey: types.StringNull(), APIKeyWO: apiKeyWO, Headers: types.ListNull(authHeadersElementType(ctx)),
		}
		return m
	}

	plan, config := planAndConfig(ctx, t, build(types.StringNull(), types.StringNull()), build(types.StringValue("pw"), types.StringValue("ak")))
	require.Empty(t, modifyPlan(ctx, private, config, &plan))

	plan, config = planAndConfig(ctx, t, build(types.StringNull(), types.StringNull()), build(types.StringValue("pw"), types.StringValue("changed")))
	diags := modifyPlan(ctx, private, config, &plan)
	require.False(t, diags.HasError())
	require.Len(t, diags, 1)
	var id types.String
	require.False(t, plan.GetAttribute(ctx, path.Root("id"), &id).HasError())
	require.True(t, id.IsUnknown())
}

func TestModifyPlanSkipsCreateAndDestroy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := newResource()

	var _ resource.ResourceWithModifyPlan = r

	s := fullSchema(ctx, t).Schema
	nullRaw := tfsdkNullRaw(ctx, s)
	plan, config := planAndConfig(ctx, t, baseModel(ctx, types.StringNull()), baseModel(ctx, types.StringValue("k")))

	// create: state is null
	resp := &resource.ModifyPlanResponse{Plan: plan}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Config: config, Plan: plan, State: tfsdk.State{Schema: s, Raw: nullRaw}}, resp)
	require.False(t, resp.Diagnostics.HasError())

	// destroy: plan is null
	resp = &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Schema: s, Raw: nullRaw}}
	r.ModifyPlan(ctx, resource.ModifyPlanRequest{Config: config, Plan: tfsdk.Plan{Schema: s, Raw: nullRaw}, State: tfsdk.State{Schema: s, Raw: plan.Raw}}, resp)
	require.False(t, resp.Diagnostics.HasError())
}
