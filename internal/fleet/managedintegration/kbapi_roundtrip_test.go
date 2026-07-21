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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullyPopulatedManagedIntegrationInputJSON is a hand-built input object in
// the managed_integrations response shape (KibanaHTTPAPIsManagedIntegration
// inputs map values). It also embeds several PackagePolicy-typed-input fields
// that must not appear on managed integration wire JSON; the companion test
// below asserts they are dropped on unmarshal/re-marshal, not silently
// carried into create/update bodies.
const fullyPopulatedManagedIntegrationInputJSON = `{
	"condition": "host.os.family == 'linux'",
	"deprecated": {"description": "dep-desc", "since": "9.0.0"},
	"enabled": true,
	"vars": {"varkey": "varval"},
	"streams": {
		"cloud_security_posture.findings": {
			"condition": "data_stream.dataset == 'audit'",
			"deprecated": {"description": "stream-dep-desc", "since": "9.1.0"},
			"enabled": true,
			"var_group_selections": {"svg1": "sopt1"},
			"vars": {"streamvarkey": "streamvarval"},
			"compiled_stream": {"compiled": "package-policy-only"},
			"config": {"scfgkey": {"frozen": true, "type": "text", "value": "scfgval"}},
			"id": "stream-id-leak",
			"release": "beta"
		}
	},
	"compiled_input": {"anything": "package-policy-only"},
	"id": "input-id-leak",
	"keep_enabled": true,
	"migrate_from": "migrate-from-x",
	"name": "input-name-leak",
	"policy_template": "cspm",
	"type": "cloudbeat/cis_aws",
	"config": {"cfgkey": {"frozen": true, "type": "text", "value": "cfgval"}}
}`

const managedIntegrationInputRoundTripDoc = `{
	"id": "policy-1",
	"name": "test-policy",
	"created_at": "2024-01-01T00:00:00.000Z",
	"created_by": "elastic",
	"updated_at": "2024-01-02T00:00:00.000Z",
	"updated_by": "elastic",
	"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "Security Posture Management"},
	"inputs": {
		"cspm-cloudbeat/cis_aws": ` + fullyPopulatedManagedIntegrationInputJSON + `
	}
}`

// TestKbapiManagedIntegrationInputRoundTrip_simplifiedFieldsSurvive is a
// regression guard against kbapi regeneration breaking the JSON round trip
// from a managed_integrations GET input entry to the PUT/POST request input
// shape. create.go and update.go build request inputs from Terraform state
// (applyCreateInputs), but the response-side and request-side generated structs
// must remain JSON-tag compatible for any future read/modify/write path and for
// hand-built fixtures such as mappedFormatManagedIntegrationJSON.
func TestKbapiManagedIntegrationInputRoundTrip_simplifiedFieldsSurvive(t *testing.T) {
	item := mustManagedIntegrationFromJSON(t, managedIntegrationInputRoundTripDoc)
	in, ok := item.Inputs["cspm-cloudbeat/cis_aws"]
	require.True(t, ok)

	b, err := json.Marshal(in)
	require.NoError(t, err)

	var req kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest
	wrapper := `{"name":"n","package":{"name":"p","version":"1"},"inputs":{"cspm-cloudbeat/cis_aws":` + string(b) + `}}`
	require.NoError(t, json.Unmarshal([]byte(wrapper), &req))
	require.NotNil(t, req.Inputs)
	reqIn, ok := (*req.Inputs)["cspm-cloudbeat/cis_aws"]
	require.True(t, ok)

	require.NotNil(t, reqIn.Condition)
	assert.Equal(t, "host.os.family == 'linux'", *reqIn.Condition)
	require.NotNil(t, reqIn.Enabled)
	assert.True(t, *reqIn.Enabled)

	require.NotNil(t, reqIn.Deprecated)
	assert.Equal(t, "dep-desc", reqIn.Deprecated.Description)
	require.NotNil(t, reqIn.Deprecated.Since)
	assert.Equal(t, "9.0.0", *reqIn.Deprecated.Since)

	require.NotNil(t, reqIn.Vars)
	varStr, err := (*reqIn.Vars)["varkey"].AsKibanaHTTPAPIsCreateManagedIntegrationRequestInputsVars0()
	require.NoError(t, err)
	assert.Equal(t, "varval", varStr)

	require.NotNil(t, reqIn.Streams)
	streams := *reqIn.Streams
	stream, ok := streams["cloud_security_posture.findings"]
	require.True(t, ok)

	require.NotNil(t, stream.Condition)
	assert.Equal(t, "data_stream.dataset == 'audit'", *stream.Condition)
	require.NotNil(t, stream.Enabled)
	assert.True(t, *stream.Enabled)

	require.NotNil(t, stream.Deprecated)
	assert.Equal(t, "stream-dep-desc", stream.Deprecated.Description)

	require.NotNil(t, stream.VarGroupSelections)
	assert.Equal(t, "sopt1", (*stream.VarGroupSelections)["svg1"])

	require.NotNil(t, stream.Vars)
	sVarStr, err := (*stream.Vars)["streamvarkey"].AsKibanaHTTPAPIsCreateManagedIntegrationRequestInputsStreamsVars0()
	require.NoError(t, err)
	assert.Equal(t, "streamvarval", sVarStr)
}

// TestKbapiManagedIntegrationInputRoundTrip_packagePolicyFieldsDropped asserts
// that legacy PackagePolicy typed-input fields present in raw JSON are not
// part of the clean KibanaHTTPAPIsManagedIntegration model and therefore must
// not survive into request bodies built via JSON round trip.
func TestKbapiManagedIntegrationInputRoundTrip_packagePolicyFieldsDropped(t *testing.T) {
	item := mustManagedIntegrationFromJSON(t, managedIntegrationInputRoundTripDoc)
	in := item.Inputs["cspm-cloudbeat/cis_aws"]

	b, err := json.Marshal(in)
	require.NoError(t, err)

	var req kbapi.KibanaHTTPAPIsCreateManagedIntegrationRequest
	wrapper := `{"name":"n","package":{"name":"p","version":"1"},"inputs":{"cspm-cloudbeat/cis_aws":` + string(b) + `}}`
	require.NoError(t, json.Unmarshal([]byte(wrapper), &req))

	reqRaw, err := json.Marshal(req)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(reqRaw, &decoded))
	inputs, ok := decoded["inputs"].(map[string]any)
	require.True(t, ok)
	inDecoded, ok := inputs["cspm-cloudbeat/cis_aws"].(map[string]any)
	require.True(t, ok)

	for _, leakKey := range []string{
		"compiled_input", "id", "keep_enabled", "migrate_from", "name",
		"policy_template", "type", "config",
	} {
		_, present := inDecoded[leakKey]
		assert.False(t, present, "PackagePolicy-only input field %q must not leak into managed integration request JSON", leakKey)
	}

	streams, ok := inDecoded["streams"].(map[string]any)
	require.True(t, ok)
	stream, ok := streams["cloud_security_posture.findings"].(map[string]any)
	require.True(t, ok)
	for _, leakKey := range []string{"compiled_stream", "config", "id", "release"} {
		_, present := stream[leakKey]
		assert.False(t, present, "PackagePolicy-only stream field %q must not leak into managed integration request JSON", leakKey)
	}
}

// TestKbapiManagedIntegrationResponse_packagePolicyTopLevelFieldsIgnored ensures
// a GET fixture shaped like an old package_policy document still unmarshals into
// KibanaHTTPAPIsManagedIntegration without surfacing PackagePolicy-only top-level
// attributes on the Go struct (they are ignored by encoding/json).
func TestKbapiManagedIntegrationResponse_packagePolicyTopLevelFieldsIgnored(t *testing.T) {
	const raw = `{
		"id": "policy-1",
		"name": "test-policy",
		"created_at": "2024-01-01T00:00:00.000Z",
		"created_by": "elastic",
		"updated_at": "2024-01-02T00:00:00.000Z",
		"updated_by": "elastic",
		"package": {"name": "cloud_security_posture", "version": "3.4.0", "title": "t"},
		"inputs": {},
		"revision": 3,
		"agent_policy_id": "ap-1",
		"output_id": "default",
		"policy_ids": ["ap-1"]
	}`

	item := mustManagedIntegrationFromJSON(t, raw)
	assert.Equal(t, "policy-1", item.Id)
	assert.Equal(t, "test-policy", item.Name)

	out, err := json.Marshal(item)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(out, &decoded))
	for _, leakKey := range []string{"revision", "agent_policy_id", "output_id", "policy_ids"} {
		_, present := decoded[leakKey]
		assert.False(t, present, "PackagePolicy-only top-level field %q must not appear on KibanaHTTPAPIsManagedIntegration", leakKey)
	}
}
