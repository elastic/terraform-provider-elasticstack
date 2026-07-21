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
	"net/http"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fleetManagedIntegrationCallRecorder records whether PUT
// /api/fleet/managed_integrations/{id} was hit (updateAgentlessPolicy's
// non-short-circuit path after the full-replace update rewrite).
func fleetManagedIntegrationCallRecorder(t *testing.T) (http.Handler, *bool) {
	t.Helper()
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(_ http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			called = true
			t.Errorf("unexpected Fleet API call for a create-only-flag-only change: %s %s", r.Method, r.URL.Path)
		}
	})
	return mux, &called
}

// TestUpdateAgentlessPolicy_createOnlyFlags covers a gap found in the Task
// 10.6 self-review: spec.md's "Create" requirement ("create_dataset_templates
// sent only on create" scenario) and "Operation flags" schema section both
// state that create_dataset_templates, force, and force_delete are
// create/delete-only knobs whose post-create changes "SHALL NOT make any API
// call" -- but because none of the three is RequiresReplace, Terraform still
// invokes Update whenever one changes, and updateAgentlessPolicy used to
// unconditionally do a GET+PUT round trip regardless of what actually
// changed. onlyCreateOnlyFlagsChanged (this file) now short-circuits that.
// Subtests below deliberately do not call t.Parallel(): they build clients
// via newTopologyTestClient, which calls clearKibanaEnvOverrides -> t.Setenv,
// and t.Setenv is documented as incompatible with parallel tests (matching
// the non-parallel style already used by TestCheckDeploymentTopology in
// topology_test.go and TestCreateAgentlessPolicy_topologyGatesFleetCall
// in create_test.go, for the same reason).
func TestUpdateAgentlessPolicy_createOnlyFlags(t *testing.T) {
	newPriorAndPlan := func(t *testing.T) (prior, plan agentlessPolicyModel) {
		t.Helper()
		prior = baseTestModel(t)
		prior.PolicyID = types.StringValue("pp-1")
		prior.ID = types.StringValue("default/pp-1")
		prior.CreateDatasetTemplates = types.BoolValue(false)
		prior.Force = types.BoolValue(false)
		prior.ForceDelete = types.BoolValue(false)
		prior.SkipTopologyCheck = types.BoolValue(false)
		// Fixture reflects the actual shape a real Terraform Update plan
		// produces (schema.go): created_at carries UseStateForUnknown, so it
		// is known and equal to prior's value; updated_at deliberately has no
		// such modifier (it legitimately changes on every real Update), so
		// the framework leaves it Unknown in the plan regardless of whether
		// this particular change is create-only-flags-only. This is exactly
		// why onlyCreateOnlyFlagsChanged (update.go) excludes both fields
		// from its comparison entirely rather than depending on plan.UpdatedAt
		// happening to equal prior.UpdatedAt: an Unknown updated_at must not
		// defeat the short-circuit. See TestOnlyCreateOnlyFlagsChanged's
		// "unknown updated_at does not defeat the short-circuit" subtest for
		// the direct proof, and the acceptance test step in acc_test.go for
		// the live-Kibana proof.
		prior.CreatedAt = types.StringValue("2024-01-01T00:00:00.000Z")
		prior.UpdatedAt = types.StringValue("2024-01-02T00:00:00.000Z")

		plan = prior
		plan.CreatedAt = types.StringValue("2024-01-01T00:00:00.000Z") // carried forward via UseStateForUnknown
		plan.UpdatedAt = types.StringUnknown()                         // no plan modifier: always Unknown in a real Update plan
		return prior, plan
	}

	t.Run("create_dataset_templates alone changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.CreateDatasetTemplates = types.BoolValue(true)

		handler, called := fleetManagedIntegrationCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			WriteID: "pp-1",
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a create_dataset_templates-only change")
		require.True(t, result.SkipReadAfterWrite, "create-only-flag short-circuit must skip envelope read-after-write")
		require.True(t, result.Model.CreateDatasetTemplates.ValueBool())
		require.Equal(t, "2024-01-02T00:00:00.000Z", result.Model.UpdatedAt.ValueString(),
			"Unknown plan updated_at must be preserved from prior when skipping read-after-write")
	})

	t.Run("force and force_delete together changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.Force = types.BoolValue(true)
		plan.ForceDelete = types.BoolValue(true)

		handler, called := fleetManagedIntegrationCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			WriteID: "pp-1",
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a force/force_delete-only change")
		require.True(t, result.SkipReadAfterWrite)
		require.True(t, result.Model.Force.ValueBool())
		require.True(t, result.Model.ForceDelete.ValueBool())
	})

	t.Run("skip_topology_check alone changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.SkipTopologyCheck = types.BoolValue(true)

		handler, called := fleetManagedIntegrationCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			WriteID: "pp-1",
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a skip_topology_check-only change")
		require.True(t, result.SkipReadAfterWrite)
		require.True(t, result.Model.SkipTopologyCheck.ValueBool())
	})

	t.Run("a create-only-flag change alongside a real attribute change still calls the API", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.CreateDatasetTemplates = types.BoolValue(true)
		plan.Description = types.StringValue("a new description")

		called := false
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				return
			}
			called = true
			http.Error(w, "not implemented in this test", http.StatusNotImplemented)
		})
		client := newTopologyTestClient(t, mux)

		_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			WriteID: "pp-1",
			SpaceID: "default",
		})

		require.True(t, diags.HasError(), "expected an error since the test server doesn't implement a real response")
		require.True(t, called, "a real attribute change alongside a create-only-flag change must still call the Fleet API")
	})
}

func TestOnlyCreateOnlyFlagsChanged(t *testing.T) {
	t.Parallel()

	// baseTestModel (not a hand-rolled struct literal) so every attr.Value
	// field is a properly-initialized null/known value rather than a
	// zero-value Go struct -- several fields here (Package, VarsJSON,
	// Inputs, CloudConnector, ...) are custom or collection types whose
	// zero value does not necessarily compare Equal to itself.
	base := baseTestModel(t)
	base.CreateDatasetTemplates = types.BoolValue(false)
	base.Force = types.BoolValue(false)
	base.ForceDelete = types.BoolValue(false)
	// created_at carries UseStateForUnknown (schema.go), so a real Update plan
	// carries it forward from prior state as a known value equal to prior's.
	// updated_at deliberately does NOT (see schema.go's comment: it
	// legitimately changes on every real Update, so promising it won't would
	// be wrong), so a real Update plan leaves it Unknown regardless of
	// whether this is a create-only-flags-only change. onlyCreateOnlyFlagsChanged
	// must not depend on either field's plan value at all -- see its doc
	// comment -- which the "unknown updated_at does not defeat the
	// short-circuit" subtest below proves directly.
	base.CreatedAt = types.StringValue("2024-01-01T00:00:00.000Z")
	base.UpdatedAt = types.StringValue("2024-01-02T00:00:00.000Z")

	t.Run("identical models", func(t *testing.T) {
		t.Parallel()
		require.True(t, onlyCreateOnlyFlagsChanged(base, base))
	})

	t.Run("only create-only flags differ", func(t *testing.T) {
		t.Parallel()
		plan := base
		plan.CreateDatasetTemplates = types.BoolValue(true)
		plan.Force = types.BoolValue(true)
		plan.ForceDelete = types.BoolValue(true)
		plan.SkipTopologyCheck = types.BoolValue(true)
		require.True(t, onlyCreateOnlyFlagsChanged(base, plan))
	})

	t.Run("a non-create-only field also differs", func(t *testing.T) {
		t.Parallel()
		plan := base
		plan.CreateDatasetTemplates = types.BoolValue(true)
		plan.Name = types.StringValue("renamed")
		require.False(t, onlyCreateOnlyFlagsChanged(base, plan))
	})

	// Direct regression proof for the real-world bug this function's
	// exclusion of created_at/updated_at fixes: updated_at has no
	// UseStateForUnknown plan modifier (schema.go), so a real Terraform
	// Update plan leaves plan.UpdatedAt Unknown even when only a create-only
	// flag changed. Before this fix, onlyCreateOnlyFlagsChanged compared
	// prior.UpdatedAt.Equal(plan.UpdatedAt) directly, which is always false
	// when plan.UpdatedAt is Unknown -- permanently defeating the
	// short-circuit outside of unit tests that hand-built a plan with a
	// matching *known* timestamp (i.e. tests that could never have caught
	// this in the first place). Since the fix removes created_at/updated_at
	// from the comparison entirely, an Unknown updated_at here must NOT
	// defeat the short-circuit. See acc_test.go's "update_flag_only" step for
	// the live-Kibana proof that this holds end to end.
	t.Run("unknown updated_at does not defeat the short-circuit", func(t *testing.T) {
		t.Parallel()
		plan := base
		plan.CreateDatasetTemplates = types.BoolValue(true)
		plan.CreatedAt = types.StringValue("2024-01-01T00:00:00.000Z") // carried forward via UseStateForUnknown
		plan.UpdatedAt = types.StringUnknown()                         // no plan modifier: realistically Unknown
		require.True(t, onlyCreateOnlyFlagsChanged(base, plan),
			"an Unknown updated_at in the plan must not defeat the short-circuit: it is server-Computed and "+
				"can never be what the user actually changed")
	})
}

// TestOnlyCreateOnlyFlagsChanged_FieldCoverage guards against
// onlyCreateOnlyFlagsChanged's allowlist silently going stale as
// agentlessPolicyModel evolves. That function is a positive allowlist (every
// field it compares is named explicitly, see its doc comment) rather than a
// negative denylist, so a new schema field added to the model in the future
// and never wired into either onlyCreateOnlyFlagsChanged's comparison chain
// or its doc comment's exclusion list would be silently treated as inert
// (Update would skip the API call even though the new field changed) instead
// of failing safe.
//
// This test uses reflection purely to enumerate agentlessPolicyModel's field
// names -- it does not attempt to verify onlyCreateOnlyFlagsChanged's
// comparison logic is behaviorally correct for each one (see
// TestOnlyCreateOnlyFlagsChanged above for that) -- and asserts every field
// appears in exactly one of the two lists below, which mirror
// onlyCreateOnlyFlagsChanged's `&&` chain and its doc comment's exclusion
// list respectively. Adding a field to the struct without adding it to one
// of these two lists (and, in the "compared" case, to the real function)
// fails this test.
func TestOnlyCreateOnlyFlagsChanged_FieldCoverage(t *testing.T) {
	t.Parallel()

	// Mirrors onlyCreateOnlyFlagsChanged's `&&` chain, field for field.
	compared := map[string]bool{
		"ID":                               true,
		"PolicyID":                         true,
		"Name":                             true,
		"Description":                      true,
		"Namespace":                        true,
		"SpaceIDs":                         true,
		"Package":                          true,
		"PolicyTemplate":                   true,
		"VarsJSON":                         true,
		"VarGroupSelections":               true,
		"Inputs":                           true,
		"CloudConnector":                   true,
		"GlobalDataTags":                   true,
		"AdditionalDatastreamsPermissions": true,
	}

	// Mirrors onlyCreateOnlyFlagsChanged's doc comment: create/delete-only
	// flags never read back from the API, provider-side plumbing that is
	// never part of the Fleet request body, and created_at/updated_at (purely
	// server-Computed -- never Optional -- so they can never be what a user's
	// config actually changed; see onlyCreateOnlyFlagsChanged's doc comment
	// for why comparing them was the original bug this exclusion fixes).
	// ResourceTimeoutsField (the embedded field itself) and Timeouts (its one
	// promoted field) are both listed since reflect.VisibleFields below
	// includes promoted fields.
	excluded := map[string]bool{
		"CreateDatasetTemplates": true,
		"Force":                  true,
		"ForceDelete":            true,
		"SkipTopologyCheck":      true,
		"KibanaConnection":       true,
		"ResourceTimeoutsField":  true, // embedded entitycore.ResourceTimeoutsField
		"Timeouts":               true, // ResourceTimeoutsField's one promoted field
		"CreatedAt":              true,
		"UpdatedAt":              true,
	}

	for _, field := range reflect.VisibleFields(reflect.TypeFor[agentlessPolicyModel]()) {
		name := field.Name
		inCompared, inExcluded := compared[name], excluded[name]
		if inCompared == inExcluded { // covers "neither" (false==false) and "both" (true==true)
			t.Errorf(
				"agentlessPolicyModel field %q is not accounted for exactly once between "+
					"onlyCreateOnlyFlagsChanged's comparison chain and its doc comment's exclusion list "+
					"(in compared list: %v, in excluded list: %v). Add it to onlyCreateOnlyFlagsChanged's "+
					"comparison chain (and this test's `compared` map) if it should trigger an API call when "+
					"changed, or to the doc comment's exclusion list (and this test's `excluded` map) if it is "+
					"client-only/provider-side plumbing that never reaches the Fleet API.",
				name, inCompared, inExcluded,
			)
		}
	}
}

func TestBuildUpdateBody_inPlaceNameAndVersion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	plan := prior
	plan.Name = types.StringValue("renamed-policy")
	pkgObj, diags := types.ObjectValueFrom(ctx, packageAttrTypes(), packageModel{
		Name:    types.StringValue("cloud_security_posture"),
		Version: types.StringValue("3.5.0"),
		Title:   types.StringValue("Security Posture Management"),
	})
	require.False(t, diags.HasError())
	plan.Package = pkgObj

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	assert.Equal(t, "renamed-policy", decoded["name"])
	pkg, ok := decoded["package"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "3.5.0", pkg["version"])
}

func TestBuildUpdateBody_cloudConnectorFromPriorState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	ccObj, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Enabled:          types.BoolValue(true),
		CloudConnectorID: types.StringValue("cc-from-state"),
		Name:             types.StringValue("write-only-name"),
		TargetCSP:        types.StringValue("aws"),
	})
	require.False(t, diags.HasError())

	prior := baseTestModel(t)
	prior.CloudConnector = ccObj

	plan := prior
	// Plan cloud_connector differs on write-only fields; PUT must still use prior.
	planCC, diags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Enabled:          types.BoolValue(false),
		CloudConnectorID: types.StringValue("cc-from-plan"),
		Name:             types.StringValue("other-name"),
		TargetCSP:        types.StringValue("gcp"),
	})
	require.False(t, diags.HasError())
	plan.CloudConnector = planCC

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	cc, ok := decoded["cloud_connector"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, cc["enabled"])
	assert.Equal(t, "cc-from-state", cc["cloud_connector_id"])
	_, hasName := cc["name"]
	_, hasTarget := cc["target_csp"]
	assert.False(t, hasName, "update must not send cloud_connector.name")
	assert.False(t, hasTarget, "update must not send cloud_connector.target_csp")
}

func TestBuildUpdateBody_fullReplaceOmitsCreateOnlyFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	prior.Force = types.BoolValue(true)
	prior.CreateDatasetTemplates = types.BoolValue(true)
	plan := prior
	plan.Force = types.BoolValue(false)
	plan.CreateDatasetTemplates = types.BoolValue(false)

	body, diags := buildUpdateBody(ctx, plan, prior)
	require.False(t, diags.HasError(), "%v", diags)

	decoded := decodeRequestJSON(t, body)
	_, hasForce := decoded["force"]
	_, hasCreateDS := decoded["create_dataset_templates"]
	_, hasID := decoded["id"]
	assert.False(t, hasForce)
	assert.False(t, hasCreateDS)
	assert.False(t, hasID)
}

func TestUpdateAgentlessPolicy_nilPrior(t *testing.T) {
	plan := baseTestModel(t)
	plan.PolicyID = types.StringValue("pp-1")
	plan.ID = types.StringValue("default/pp-1")

	var unexpectedCalls atomic.Int64
	client := newTopologyTestClient(t, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		unexpectedCalls.Add(1)
	}))

	_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
		Plan:    plan,
		Prior:   nil,
		WriteID: "pp-1",
		SpaceID: "default",
	})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "missing prior state")
	require.Equal(t, int64(0), unexpectedCalls.Load())
}

func TestUpdateAgentlessPolicy_putError(t *testing.T) {
	prior := baseTestModel(t)
	prior.PolicyID = types.StringValue("pp-1")
	prior.ID = types.StringValue("default/pp-1")
	plan := prior
	plan.Description = types.StringValue("changed")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			http.Error(w, "boom", http.StatusInternalServerError)
		}
	})
	client := newTopologyTestClient(t, mux)

	_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: "pp-1",
		SpaceID: "default",
	})
	require.True(t, diags.HasError())
}

func TestUpdateAgentlessPolicy_notFoundOnPut(t *testing.T) {
	prior := baseTestModel(t)
	prior.PolicyID = types.StringValue("pp-1")
	prior.ID = types.StringValue("default/pp-1")
	plan := prior
	plan.Description = types.StringValue("changed")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
		}
	})
	client := newTopologyTestClient(t, mux)

	_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: "pp-1",
		SpaceID: "default",
	})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "Unexpected status code")
}

func TestUpdateAgentlessPolicy_successDoesNotSkipReadAfterWrite(t *testing.T) {
	prior := baseTestModel(t)
	prior.PolicyID = types.StringValue("pp-1")
	prior.ID = types.StringValue("default/pp-1")
	plan := prior
	plan.Description = types.StringValue("changed")

	mux := http.NewServeMux()
	legacyCalls := registerLegacyPackagePoliciesGuard(mux)
	method := newHTTPMethodCapture()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			return
		}
		method.record(r)
		w.Header().Set("Content-Type", "application/json")
		const putOK = `{"item":{"id":"pp-1","name":"test-policy","package":{"name":"cloud_security_posture",` +
			`"version":"3.4.0"},"created_at":"2024-01-01T00:00:00.000Z","updated_at":"2024-01-02T00:00:00.000Z"}}`
		_, _ = w.Write([]byte(putOK))
	})
	client := newTopologyTestClient(t, mux)

	result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: "pp-1",
		SpaceID: "default",
	})
	require.False(t, diags.HasError(), "%v", diags)
	method.requireEqual(t, http.MethodPut)
	requireNoLegacyPackagePoliciesCalls(t, legacyCalls)
	require.False(t, result.SkipReadAfterWrite, "real PUT must leave read-after-write to the envelope")
}

func TestUpdateAgentlessPolicy_malformedPutResponseBody(t *testing.T) {
	prior := baseTestModel(t)
	prior.PolicyID = types.StringValue("pp-1")
	prior.ID = types.StringValue("default/pp-1")
	plan := prior
	plan.Description = types.StringValue("changed")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusOK)
		}
	})
	client := newTopologyTestClient(t, mux)

	_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
		Plan:    plan,
		Prior:   &prior,
		WriteID: "pp-1",
		SpaceID: "default",
	})
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Summary(), "Failed to parse response")
}

func TestBuildUpdateBody_cloudConnectorOmittedWhenPriorUnset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baseTestModel(t)
	plan := prior
	plan.Description = types.StringValue("x")

	body, diags := buildUpdateBody(ctx, plan, prior)
	require.False(t, diags.HasError(), "%v", diags)
	decoded := decodeRequestJSON(t, body)
	_, present := decoded["cloud_connector"]
	assert.False(t, present)
}

func TestBuildUpdateBody_omitsKnownNullOptionalFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baseTestModel(t)
	plan := prior
	plan.Description = types.StringNull()
	plan.Namespace = types.StringNull()
	plan.PolicyTemplate = types.StringNull()
	plan.Inputs = policyshape.NewInputsNull(agentlessInputType())
	plan.VarGroupSelections = types.MapNull(types.StringType)
	plan.AdditionalDatastreamsPermissions = types.ListNull(types.StringType)
	plan.GlobalDataTags = types.MapNull(globalDataTagsElementType())

	body, diags := buildUpdateBody(ctx, plan, prior)
	require.False(t, diags.HasError(), "%v", diags)
	decoded := decodeRequestJSON(t, body)
	_, hasDesc := decoded["description"]
	_, hasNS := decoded["namespace"]
	_, hasPT := decoded["policy_template"]
	_, hasInputs := decoded["inputs"]
	_, hasVGS := decoded["var_group_selections"]
	_, hasPerms := decoded["additional_datastreams_permissions"]
	_, hasTags := decoded["global_data_tags"]
	assert.False(t, hasDesc)
	assert.False(t, hasNS)
	assert.False(t, hasPT)
	assert.False(t, hasInputs)
	assert.False(t, hasVGS)
	assert.False(t, hasPerms)
	assert.False(t, hasTags)
}

func TestBuildUpdateBody_unknownTopLevelFieldsErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	basePrior := func(t *testing.T) agentlessPolicyModel {
		t.Helper()
		return baseTestModel(t)
	}

	tests := []struct {
		name   string
		mutate func(*agentlessPolicyModel)
	}{
		{"description", func(m *agentlessPolicyModel) { m.Description = types.StringUnknown() }},
		{"namespace", func(m *agentlessPolicyModel) { m.Namespace = types.StringUnknown() }},
		{"policy_template", func(m *agentlessPolicyModel) { m.PolicyTemplate = types.StringUnknown() }},
		{"vars_json", func(m *agentlessPolicyModel) { m.VarsJSON = policyshape.NewVarsJSONUnknown() }},
		{"var_group_selections", func(m *agentlessPolicyModel) { m.VarGroupSelections = types.MapUnknown(types.StringType) }},
		{"additional_datastreams_permissions", func(m *agentlessPolicyModel) {
			m.AdditionalDatastreamsPermissions = types.ListUnknown(types.StringType)
		}},
		{"global_data_tags", func(m *agentlessPolicyModel) { m.GlobalDataTags = types.MapUnknown(globalDataTagsElementType()) }},
		{"package", func(m *agentlessPolicyModel) { m.Package = types.ObjectUnknown(packageAttrTypes()) }},
		{"inputs", func(m *agentlessPolicyModel) {
			m.Inputs = policyshape.InputsValue{MapValue: types.MapUnknown(policyshape.NewInputsType(agentlessInputType()))}
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prior := basePrior(t)
			plan := prior
			tc.mutate(&plan)
			_, diags := buildUpdateBody(ctx, plan, prior)
			require.True(t, diags.HasError())
			require.Contains(t, diags.Errors()[0].Detail(), "attribute is unknown")
		})
	}
}

func TestBuildUpdateBody_unknownInputsErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baseTestModel(t)
	plan := prior
	plan.Inputs = policyshape.InputsValue{
		MapValue: types.MapUnknown(policyshape.NewInputsType(agentlessInputType())),
	}

	_, diags := buildUpdateBody(ctx, plan, prior)
	require.True(t, diags.HasError())
	require.Contains(t, diags.Errors()[0].Detail(), "inputs attribute is unknown")
}

func TestBuildUpdateBody_fullReplaceExtraFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baseTestModel(t)
	plan := prior
	plan.PolicyTemplate = types.StringValue("cspm")

	vgs, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{"vg1": "opt-a"})
	require.False(t, diags.HasError())
	plan.VarGroupSelections = vgs

	plan.AdditionalDatastreamsPermissions, diags = types.ListValueFrom(ctx, types.StringType, []string{"logs-*"})
	require.False(t, diags.HasError())

	plan.GlobalDataTags, diags = types.MapValueFrom(ctx, globalDataTagsElementType(), map[string]attr.Value{
		"cost": types.ObjectValueMust(globalDataTagAttrTypes(), map[string]attr.Value{
			globalDataTagStringValueAttr: types.StringNull(),
			globalDataTagNumberValueAttr: types.Float32Value(42),
		}),
	})
	require.False(t, diags.HasError())

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)
	decoded := decodeRequestJSON(t, body)
	assert.Equal(t, "cspm", decoded["policy_template"])

	vgsOut, ok := decoded["var_group_selections"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "opt-a", vgsOut["vg1"])

	perms, ok := decoded["additional_datastreams_permissions"].([]any)
	require.True(t, ok)
	require.Len(t, perms, 1)

	tags, ok := decoded["global_data_tags"].([]any)
	require.True(t, ok)
	require.Len(t, tags, 1)
	tag0, ok := tags[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cost", tag0["name"])
	assert.InDelta(t, float64(42), tag0["value"], 0.001)
}

func TestBuildUpdateBody_knownEmptyInputsMapSendsEmptyObject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	prior := baseTestModel(t)
	plan := prior
	emptyInputs, diags := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{})
	require.False(t, diags.HasError())
	plan.Inputs = emptyInputs

	body, bodyDiags := buildUpdateBody(ctx, plan, prior)
	require.False(t, bodyDiags.HasError(), "%v", bodyDiags)

	decoded := decodeRequestJSON(t, body)
	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok, "full-replace update must send inputs when the plan map is known-empty")
	assert.Empty(t, inputs)
}

func TestBuildUpdateBody_sendsEmptyVarGroupSelectionsMap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	prior := baseTestModel(t)
	plan := prior
	plan.VarGroupSelections, _ = types.MapValueFrom(ctx, types.StringType, map[string]string{})

	body, diags := buildUpdateBody(ctx, plan, prior)
	require.False(t, diags.HasError(), "%v", diags)
	decoded := decodeRequestJSON(t, body)
	vgs, ok := decoded["var_group_selections"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, vgs)
}
