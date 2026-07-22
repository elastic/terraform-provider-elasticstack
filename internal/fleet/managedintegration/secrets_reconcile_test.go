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

func TestReconcileSecretVarsMapFromPrior_bareRef(t *testing.T) {
	t.Parallel()
	prior := map[string]any{"token": "plaintext-secret"}
	resp := map[string]any{"token": map[string]any{"id": "ref-1", "isSecretRef": true}}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, "plaintext-secret", resp["token"])
}

func TestReconcileSecretVarsMapFromPrior_wrappedRef(t *testing.T) {
	t.Parallel()
	prior := map[string]any{"token": "plaintext-secret"}
	resp := map[string]any{"token": map[string]any{"value": map[string]any{"id": "ref-2", "isSecretRef": true}}}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, "plaintext-secret", resp["token"])
}

func TestReconcileSecretVarsMapFromPrior_keepsRefOnImport(t *testing.T) {
	t.Parallel()
	resp := map[string]any{"token": map[string]any{"id": "ref-1", "isSecretRef": true}}
	reconcileSecretVarsMapFromPrior(nil, resp)
	assert.Equal(t, map[string]any{"id": "ref-1", "isSecretRef": true}, resp["token"])
}

func TestReconcileSecretVarsMapFromPrior_unchangedPlaintext(t *testing.T) {
	t.Parallel()
	prior := map[string]any{"region": "us-east-1"}
	resp := map[string]any{"region": "us-east-1"}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, "us-east-1", resp["region"])
}

func TestReconcileManagedIntegrationSecretsFromPrior_streamVars(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	streamsPrior, d := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.credentials.external_id":"my-plaintext"}`),
		},
	})
	require.False(t, d.HasError())

	priorInputs, d := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: streamsPrior,
		},
	})
	require.False(t, d.HasError())

	streamsPop, d := types.MapValueFrom(ctx, policyshape.StreamType(), map[string]policyshape.InputStreamModel{
		"cloud_security_posture.findings": {
			Enabled: types.BoolValue(true),
			Vars:    jsontypes.NewNormalizedValue(`{"aws.credentials.external_id":{"id":"secret-id","isSecretRef":true}}`),
		},
	})
	require.False(t, d.HasError())

	popInputs, d := policyshape.NewInputsValueFrom(ctx, agentlessInputType(), map[string]agentlessInputModel{
		"cspm-cloudbeat/cis_aws": {
			Enabled: types.BoolValue(true),
			Streams: streamsPop,
		},
	})
	require.False(t, d.HasError())

	prior := agentlessPolicyModel{Inputs: priorInputs}
	populated := agentlessPolicyModel{Inputs: popInputs}

	var diags diag.Diagnostics
	reconcileManagedIntegrationSecretsFromPrior(ctx, &prior, &populated, &diags)
	require.False(t, diags.HasError())

	inputsMap := typeutils.MapTypeAs[agentlessInputModel](ctx, populated.Inputs.MapValue, path.Root(attrInputs), &diags)
	require.False(t, diags.HasError())
	streamsMap := typeutils.MapTypeAs[policyshape.InputStreamModel](ctx, inputsMap["cspm-cloudbeat/cis_aws"].Streams, path.Root(attrInputs), &diags)
	require.False(t, diags.HasError())
	assert.Contains(t, streamsMap["cloud_security_posture.findings"].Vars.ValueString(), "my-plaintext")
}

func TestReconcileSecretVarsMapFromPrior_multiIDRefs(t *testing.T) {
	t.Parallel()
	prior := map[string]any{"hosts": []any{"secret-a", "secret-b"}}
	resp := map[string]any{"hosts": map[string]any{"isSecretRef": true, "ids": []any{"id-a", "id-b"}}}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, []any{"secret-a", "secret-b"}, resp["hosts"])
}

func TestReconcileSecretVarsMapFromPrior_multiIDMissingPriorKeepsRef(t *testing.T) {
	t.Parallel()
	resp := map[string]any{"hosts": map[string]any{"isSecretRef": true, "ids": []any{"id-a", "id-b"}}}
	reconcileSecretVarsMapFromPrior(map[string]any{}, resp)
	assert.Equal(t, map[string]any{"isSecretRef": true, "ids": []any{"id-a", "id-b"}}, resp["hosts"])
}

func TestReconcileSecretVarsMapFromPrior_multiIDWrappedRef(t *testing.T) {
	t.Parallel()
	prior := map[string]any{"token": "plain"}
	resp := map[string]any{"token": map[string]any{"value": map[string]any{"isSecretRef": true, "id": "one-id"}}}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, "plain", resp["token"])
}

func TestReconcileSecretVarsMapFromPrior_priorSecretRefUnchanged(t *testing.T) {
	t.Parallel()
	ref := map[string]any{"isSecretRef": true, "id": "same"}
	prior := map[string]any{"token": ref}
	resp := map[string]any{"token": map[string]any{"isSecretRef": true, "id": "other"}}
	reconcileSecretVarsMapFromPrior(prior, resp)
	assert.Equal(t, ref, resp["token"])
}

func TestPopulateFromManagedIntegration_cloudConnectorFromAPIOnImport(t *testing.T) {
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
		"cloud_connector": {"enabled": true, "cloud_connector_id": "cc-import"}
	}`)

	m := agentlessPolicyModel{}
	popDiags := m.populateFromManagedIntegration(ctx, "default", item, nil)
	require.False(t, popDiags.HasError())

	var cc cloudConnectorModel
	require.False(t, m.CloudConnector.As(ctx, &cc, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "cc-import", cc.CloudConnectorID.ValueString())
	assert.True(t, cc.Enabled.ValueBool())
	assert.True(t, cc.Name.IsNull())
	assert.True(t, cc.TargetCSP.IsNull())
}
