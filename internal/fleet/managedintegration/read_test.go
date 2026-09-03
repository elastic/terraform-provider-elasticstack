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
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadManagedIntegration_notFound signals removed out-of-band without error.
func TestReadManagedIntegration_notFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"statusCode":404,"error":"Not Found","message":"managed integration not found"}`)
	})
	client := newTopologyTestClient(t, mux)

	seeded := baseTestModel(t)
	seeded.PolicyID = types.StringValue("policy-1")
	seeded.Name = types.StringValue("seed-name-must-not-change")

	out, ok, diags := readManagedIntegration(context.Background(), client, "policy-1", "default", seeded)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, ok)
	assert.Equal(t, "seed-name-must-not-change", out.Name.ValueString())
}

// TestReadManagedIntegration_serverError propagates GET failures.
func TestReadManagedIntegration_serverError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"statusCode":500,"error":"Internal Server Error","message":"boom"}`)
	})
	client := newTopologyTestClient(t, mux)

	seeded := baseTestModel(t)
	seeded.PolicyID = types.StringValue("policy-1")

	_, ok, diags := readManagedIntegration(context.Background(), client, "policy-1", "default", seeded)
	require.True(t, diags.HasError())
	require.False(t, ok)
}

// TestReadManagedIntegration_populatesFromManagedIntegration exercises GET
// /api/fleet/managed_integrations/{id} (ReadManagedIntegration) and
// populateFromManagedIntegration while preserving create-only flags from the
// incoming model.
func TestReadManagedIntegration_populatesFromManagedIntegration(t *testing.T) {
	const managedIntegrationJSON = `{"item":{` +
		`"id":"policy-1","name":"api-name",` +
		`"created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic",` +
		`"updated_at":"2026-01-02T00:00:00.000Z","updated_by":"elastic",` +
		`"inputs":{},` +
		`"package":{"name":"cloud_security_posture","version":"3.4.0"},` +
		`"cloud_connector":{"enabled":true,"cloud_connector_id":"cc-from-api"}` +
		`}}`

	mux := http.NewServeMux()
	legacyCalls := registerLegacyPackagePoliciesGuard(mux)
	method := newHTTPMethodCapture()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		method.record(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, managedIntegrationJSON)
	})
	client := newTopologyTestClient(t, mux)

	ctx := context.Background()
	ccObj, ccDiags := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cloudConnectorModel{
		Name:             types.StringValue("seed-connector-name"),
		TargetCSP:        types.StringValue("aws"),
		Enabled:          types.BoolValue(true),
		CloudConnectorID: types.StringValue("cc-stale-in-state"),
	})
	require.False(t, ccDiags.HasError())

	seeded := baseTestModel(t)
	seeded.Force = types.BoolValue(true)
	seeded.CreateDatasetTemplates = types.BoolValue(true)
	seeded.SkipTopologyCheck = types.BoolValue(true)
	seeded.ForceDelete = types.BoolValue(true)
	seeded.CloudConnector = ccObj

	out, ok, diags := readManagedIntegration(context.Background(), client, "policy-1", "default", seeded)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, ok)
	method.requireEqual(t, http.MethodGet)
	requireNoLegacyPackagePoliciesCalls(t, legacyCalls)
	assert.Equal(t, "api-name", out.Name.ValueString())
	assert.Equal(t, "default/policy-1", out.ID.ValueString())
	assert.True(t, out.Force.ValueBool())
	assert.True(t, out.CreateDatasetTemplates.ValueBool())
	assert.True(t, out.SkipTopologyCheck.ValueBool())
	assert.True(t, out.ForceDelete.ValueBool())

	var cc cloudConnectorModel
	require.False(t, out.CloudConnector.As(ctx, &cc, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "seed-connector-name", cc.Name.ValueString())
	assert.Equal(t, "aws", cc.TargetCSP.ValueString())
	assert.Equal(t, "cc-from-api", cc.CloudConnectorID.ValueString())
	assert.True(t, cc.Enabled.ValueBool())
}

// TestReadManagedIntegration_callback_preservesStreamSecretVars wires readManagedIntegration
// end-to-end: GET returns kbapi managed-integration stream vars as a Fleet secret
// reference; prior state holds plaintext; populate + reconcile must keep plaintext.
func TestReadManagedIntegration_callback_preservesStreamSecretVars(t *testing.T) {
	const plaintextExternalID = "my-plaintext-external-id"
	const managedIntegrationJSON = `{"item":{` +
		`"id":"policy-1","name":"api-name",` +
		`"created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic",` +
		`"updated_at":"2026-01-02T00:00:00.000Z","updated_by":"elastic",` +
		`"inputs":{` +
		`"cspm-cloudbeat/cis_aws":{` +
		`"enabled":true,` +
		`"streams":{` +
		`"cloud_security_posture.findings":{` +
		`"enabled":true,` +
		`"vars":{` +
		`"aws.credentials.external_id":{"id":"secret-from-api","isSecretRef":true}` +
		`}` +
		`}` +
		`}` +
		`}` +
		`},` +
		`"package":{"name":"cloud_security_posture","version":"3.4.0"}` +
		`}}`

	mux := http.NewServeMux()
	legacyCalls := registerLegacyPackagePoliciesGuard(mux)
	method := newHTTPMethodCapture()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, r *http.Request) {
		method.record(r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, managedIntegrationJSON)
	})
	client := newTopologyTestClient(t, mux)

	ctx := context.Background()
	streamsPrior, d := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.credentials.external_id":` + fmt.Sprintf("%q", plaintextExternalID) + `}`),
		},
	})
	require.False(t, d.HasError())

	priorInputs, d := policyshape.NewInputsValueFrom(ctx, managedIntegrationInputType(), map[string]managedIntegrationInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: streamsPrior,
		},
	})
	require.False(t, d.HasError())

	seeded := baseTestModel(t)
	seeded.PolicyID = types.StringValue("policy-1")
	seeded.Inputs = priorInputs

	out, ok, diags := readManagedIntegration(ctx, client, "policy-1", "default", seeded)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, ok)
	method.requireEqual(t, http.MethodGet)
	requireNoLegacyPackagePoliciesCalls(t, legacyCalls)

	var assertDiags diag.Diagnostics
	inputsMap := typeutils.MapTypeAs[managedIntegrationInputModel](ctx, out.Inputs.MapValue, path.Root(attrInputs), &assertDiags)
	require.False(t, assertDiags.HasError())
	streamsMap := typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, inputsMap["cspm-cloudbeat/cis_aws"].Streams, path.Root(attrInputs), &assertDiags)
	require.False(t, assertDiags.HasError())
	assert.Contains(t, streamsMap["cloud_security_posture.findings"].Vars.ValueString(), plaintextExternalID)
	assert.NotContains(t, streamsMap["cloud_security_posture.findings"].Vars.ValueString(), "isSecretRef")
}
