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
	"reflect"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/providerfwtest"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestResource_embedsKibanaResource(t *testing.T) {
	t.Parallel()
	rt := reflect.TypeFor[Resource]()
	field, ok := rt.FieldByName("KibanaResource")
	require.True(t, ok)
	require.True(t, field.Anonymous)
	require.Equal(t, reflect.TypeFor[*entitycore.KibanaResource[agentlessPolicyModel]](), field.Type)
}

func TestResource_importState_customCompositeID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)
	st := providerfwtest.EmptyImportState(t, r)
	resp := &resource.ImportStateResponse{State: st}

	r.ImportState(ctx, resource.ImportStateRequest{ID: "myspace/my-policy-id"}, resp)
	require.False(t, resp.Diagnostics.HasError())

	var policyID types.String
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("policy_id"), &policyID)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Equal(t, "my-policy-id", policyID.ValueString())

	var spaceIDs types.Set
	resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_ids"), &spaceIDs)...)
	require.False(t, resp.Diagnostics.HasError())

	var elems []types.String
	resp.Diagnostics.Append(spaceIDs.ElementsAs(ctx, &elems, false)...)
	require.False(t, resp.Diagnostics.HasError())
	require.Len(t, elems, 1)
	require.Equal(t, "myspace", elems[0].ValueString())
}

func TestAgentlessPolicyModel_getVersionRequirements(t *testing.T) {
	t.Parallel()

	m := agentlessPolicyModel{}
	reqs, diags := m.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.True(t, reqs[0].MinVersion.Equal(version.Must(version.NewVersion("9.5.0"))))
	require.Equal(t, "Fleet managed integrations require Elastic Stack v9.5.0 or later (experimental API).", reqs[0].ErrorMessage)
}

func TestMinVersion_matchesPolicyshapeMinVersionCondition(t *testing.T) {
	t.Parallel()
	require.True(t, MinVersion.Equal(policyshape.MinVersionCondition),
		"resource MinVersion and policyshape.MinVersionCondition must stay aligned so the envelope gate guarantees `condition` support")
}

// fakeMinVersionClient is a minimal stand-in for a *clients.KibanaScopedClient
// that only satisfies entitycore.MinVersionClient. It records whether/how it
// was called so tests can assert the version gate actually consulted it
// (rather than, say, vacuously passing because GetVersionRequirements
// returned no requirements).
type fakeMinVersionClient struct {
	supported           bool
	called              bool
	requestedMinVersion *version.Version
}

func (f *fakeMinVersionClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	f.called = true
	f.requestedMinVersion = minVersion
	return f.supported, nil
}

// TestAgentlessPolicyModel_versionGate_firesBeforeAPICall is Task 6.1's test:
// it asserts that the version check fires before any API call is attempted.
//
// entitycore.EnforceVersionRequirements is the exact function
// kibana_resource_envelope.go's Create/Read/Update/Delete call -- before
// invoking createAgentlessPolicy/read/update/delete -- whenever the decoded
// model satisfies entitycore.WithVersionRequirements (see
// internal/entitycore/kibana_resource_envelope.go and
// internal/entitycore/version_requirements.go). That generic short-circuit
// behavior (the callback is never invoked when this function returns error
// diagnostics) is already covered exhaustively by
// TestKibanaResource_Create_versionReqDiagsStopCreate and its
// Read/Update/Delete siblings in
// internal/entitycore/kibana_resource_envelope_test.go, using a synthetic
// model. What those generic tests do NOT cover is whether *this* resource's
// GetVersionRequirements (agentlessPolicyModel, MinVersion 9.5.0) actually
// causes that gate to fire against an unsupported Kibana -- that is what
// this test proves, using the resource's real, production
// agentlessPolicyModel and the real entitycore.EnforceVersionRequirements
// function.
//
// Together, these two facts (the envelope never calls the API when this
// function errors; this function does error for agentlessPolicyModel against
// a sub-9.5.0 Kibana) establish that Create/Read/Update/Delete on
// elasticstack_fleet_managed_integration never reach the Fleet API when the
// connected Kibana is older than 9.5.0. See
// openspec/changes/fleet-managed-integration/specs/fleet-managed-integration/
// spec.md, requirement "Version gate for managed_integrations endpoint" ->
// "Scenario: Older Kibana returns error".
func TestAgentlessPolicyModel_versionGate_firesBeforeAPICall(t *testing.T) {
	t.Parallel()

	t.Run("unsupported version blocks with no API call", func(t *testing.T) {
		t.Parallel()

		client := &fakeMinVersionClient{supported: false}
		apiCallMade := false

		diags := entitycore.EnforceVersionRequirements(context.Background(), client, agentlessPolicyModel{})

		// Mirrors kibana_resource_envelope.go's control flow: the real
		// Create/Read/Update/Delete callback (which is what would issue the
		// actual Fleet API call) only runs if EnforceVersionRequirements
		// produced no error diagnostics.
		if !diags.HasError() {
			apiCallMade = true
		}

		require.True(t, client.called, "EnforceMinVersion must be consulted to evaluate the version requirement")
		require.NotNil(t, client.requestedMinVersion)
		require.True(t, client.requestedMinVersion.Equal(version.Must(version.NewVersion("9.5.0"))))
		require.True(t, diags.HasError(), "version gate must produce an error diagnostic for an unsupported Kibana")
		require.False(t, apiCallMade, "no API call should be attempted when the version gate fails")

		var summaries []string
		for _, d := range diags {
			summaries = append(summaries, d.Summary()+": "+d.Detail())
		}
		require.Contains(t, strings.Join(summaries, "\n"), "9.5.0")
		require.Contains(t, strings.Join(summaries, "\n"), "Fleet managed integrations require Elastic Stack v9.5.0 or later (experimental API).")
	})

	t.Run("supported version does not block", func(t *testing.T) {
		t.Parallel()

		client := &fakeMinVersionClient{supported: true}
		diags := entitycore.EnforceVersionRequirements(context.Background(), client, agentlessPolicyModel{})

		require.True(t, client.called)
		require.False(t, diags.HasError(), "version gate must not block a supported Kibana version")
	})
}

func TestAgentlessPolicyModel_getSpaceID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("null space_ids defaults to \"default\"", func(t *testing.T) {
		t.Parallel()
		// space_ids has no schema-level Default plan modifier (matching
		// internal/fleet/output and internal/fleet/serverhost), so Create
		// must fall back to "default" here or every Create without an
		// explicit space_ids would fail entitycore's validateSpaceID check
		// (this resource is space-scoped, unlike output/serverhost, which
		// opt out via KibanaUnscopedSpace). See the GetSpaceID doc comment.
		m := agentlessPolicyModel{SpaceIDs: types.SetNull(types.StringType)}
		require.Equal(t, "default", m.GetSpaceID().ValueString())
	})

	t.Run("unknown space_ids defaults to \"default\"", func(t *testing.T) {
		t.Parallel()
		m := agentlessPolicyModel{SpaceIDs: types.SetUnknown(types.StringType)}
		require.Equal(t, "default", m.GetSpaceID().ValueString())
	})

	t.Run("returns first non-empty element", func(t *testing.T) {
		t.Parallel()
		set, diags := types.SetValueFrom(ctx, types.StringType, []string{"myspace"})
		require.False(t, diags.HasError())
		m := agentlessPolicyModel{SpaceIDs: set}
		require.Equal(t, "myspace", m.GetSpaceID().ValueString())
	})
}
