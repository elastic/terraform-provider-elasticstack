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

package agentlesspolicy

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// fleetPackagePolicyCallRecorder builds an http.Handler that records whether
// GET/PUT /api/fleet/package_policies/{id} was ever hit -- the two calls
// updateAgentlessPolicy's normal (non-short-circuit) path makes via
// fleetclient.GetDefendPackagePolicy / fleetclient.UpdateAgentlessPolicyViaPackagePolicy.
// Reused by TestUpdateAgentlessPolicy_createOnlyFlags below to prove the
// spec.md "Create" requirement's "changing create_dataset_templates after
// creation SHALL NOT make any API call" guarantee (and its force/force_delete
// analogues) actually holds through updateAgentlessPolicy's real call site,
// not just in isolation.
func fleetPackagePolicyCallRecorder(t *testing.T) (http.Handler, *bool) {
	t.Helper()
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/package_policies/", func(_ http.ResponseWriter, r *http.Request) {
		called = true
		t.Errorf("unexpected Fleet API call for a create-only-flag-only change: %s %s", r.Method, r.URL.Path)
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

		plan = prior
		return prior, plan
	}

	t.Run("create_dataset_templates alone changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.CreateDatasetTemplates = types.BoolValue(true)

		handler, called := fleetPackagePolicyCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a create_dataset_templates-only change")
		require.True(t, result.Model.CreateDatasetTemplates.ValueBool())
	})

	t.Run("force and force_delete together changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.Force = types.BoolValue(true)
		plan.ForceDelete = types.BoolValue(true)

		handler, called := fleetPackagePolicyCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a force/force_delete-only change")
		require.True(t, result.Model.Force.ValueBool())
		require.True(t, result.Model.ForceDelete.ValueBool())
	})

	t.Run("skip_topology_check alone changing makes no API call", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.SkipTopologyCheck = types.BoolValue(true)

		handler, called := fleetPackagePolicyCallRecorder(t)
		client := newTopologyTestClient(t, handler)

		result, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
			SpaceID: "default",
		})

		require.False(t, diags.HasError(), "%v", diags)
		require.False(t, *called, "no Fleet API call should be made for a skip_topology_check-only change")
		require.True(t, result.Model.SkipTopologyCheck.ValueBool())
	})

	t.Run("a create-only-flag change alongside a real attribute change still calls the API", func(t *testing.T) {
		prior, plan := newPriorAndPlan(t)
		plan.CreateDatasetTemplates = types.BoolValue(true)
		plan.Description = types.StringValue("a new description")

		called := false
		mux := http.NewServeMux()
		mux.HandleFunc("/api/fleet/package_policies/", func(w http.ResponseWriter, _ *http.Request) {
			called = true
			// Any real handling beyond "was it called" is out of scope for
			// this test; an error status is enough to prove the
			// short-circuit did not fire without needing a full fixture
			// response body.
			http.Error(w, "not implemented in this test", http.StatusNotImplemented)
		})
		client := newTopologyTestClient(t, mux)

		_, diags := updateAgentlessPolicy(context.Background(), client, entitycore.KibanaWriteRequest[agentlessPolicyModel]{
			Plan:    plan,
			Prior:   &prior,
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
		"CreatedAt":                        true,
		"UpdatedAt":                        true,
	}

	// Mirrors onlyCreateOnlyFlagsChanged's doc comment: create/delete-only
	// flags never read back from the API, plus provider-side plumbing that
	// is never part of the Fleet request body. ResourceTimeoutsField (the
	// embedded field itself) and Timeouts (its one promoted field) are both
	// listed since reflect.VisibleFields below includes promoted fields.
	excluded := map[string]bool{
		"CreateDatasetTemplates": true,
		"Force":                  true,
		"ForceDelete":            true,
		"SkipTopologyCheck":      true,
		"KibanaConnection":       true,
		"ResourceTimeoutsField":  true, // embedded entitycore.ResourceTimeoutsField
		"Timeouts":               true, // ResourceTimeoutsField's one promoted field
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
