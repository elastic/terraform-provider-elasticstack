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

// fullyPopulatedTypedInputJSON is a hand-built response-shaped JSON document
// covering every field of kbapi.PackagePolicyTypedInput (the real generated
// response-side type buildUpdateInputs decodes from
// current.Inputs.AsPackagePolicyTypedInputs()), including its nested
// PackagePolicyTypedInputStream, with every field set to a distinguishable,
// non-zero/non-default value. compiled_input is deliberately included too:
// it exists only on the response side (kbapi.PackagePolicyRequestTypedInput
// has no field for it at all) and this test's assertions below confirm it is
// -- correctly -- dropped, not silently misrouted somewhere else.
const fullyPopulatedTypedInputJSON = `{
	"compiled_input": {"anything": "response-only, must not survive the round trip"},
	"condition": "host.os.family == 'linux'",
	"config": {"cfgkey": {"frozen": true, "type": "text", "value": "cfgval"}},
	"deprecated": {"description": "dep-desc", "replaced_by": {"old": "new"}, "since": "9.0.0"},
	"enabled": true,
	"id": "input-id-1",
	"keep_enabled": true,
	"migrate_from": "migrate-from-x",
	"name": "input-name-1",
	"policy_template": "cspm",
	"type": "cloudbeat/cis_aws",
	"var_group_selections": {"vg1": "opt1"},
	"vars": {"varkey": {"frozen": false, "type": "text", "value": "varval"}},
	"streams": [
		{
			"compiled_stream": {"compiled": "yes"},
			"condition": "data_stream.dataset == 'audit'",
			"config": {"scfgkey": {"frozen": true, "type": "text", "value": "scfgval"}},
			"data_stream": {
				"dataset": "cloud_security_posture.findings",
				"type": "logs",
				"elasticsearch": {
					"dynamic_dataset": true,
					"dynamic_namespace": true,
					"privileges": {"indices": ["idx1", "idx2"]}
				}
			},
			"deprecated": {"description": "stream-dep-desc", "since": "9.1.0"},
			"enabled": true,
			"id": "stream-id-1",
			"keep_enabled": true,
			"migrate_from": "stream-migrate-from",
			"release": "beta",
			"var_group_selections": {"svg1": "sopt1"},
			"vars": {"streamvarkey": {"frozen": true, "type": "text", "value": "streamvarval"}}
		}
	]
}`

// TestKbapiTypedInputRoundTrip_allFieldsSurvive is a regression guard against
// a future kbapi regeneration silently breaking update.go's buildUpdateInputs
// round trip: for every input it echoes back on Update, buildUpdateInputs
// does exactly `b, _ := json.Marshal(in); json.Unmarshal(b, &reqIn)` to get
// from the response-side kbapi.PackagePolicyTypedInput to the request-side
// kbapi.PackagePolicyRequestTypedInput (see that function's doc comment and
// this package's header comment on why: oapi-codegen gives several of their
// shared fields, like Vars/Config, anonymous struct types that can't be
// spelled out by hand without duplicating oapi-codegen's own generated
// shape). That round trip is JSON-tag-matched, not Go-field-name-matched --
// if a future kbapi regeneration renamed or dropped a JSON tag on only one
// of the two types, this would still compile (both types are independently
// valid Go structs), but the affected field would silently stop
// round-tripping in production with no test currently catching it.
//
// This test builds a real kbapi.PackagePolicyTypedInput (via JSON, the same
// way a real Kibana response would populate it, and the only practical way
// to populate the Vars/Config anonymous-struct fields without hand-spelling
// oapi-codegen's own generated type shape) with every field set to a
// distinguishable value, runs it through that exact round trip, and asserts
// every field survived unchanged -- except compiled_input, which is
// response-only (kbapi.PackagePolicyRequestTypedInput has no field for it)
// and is asserted absent, not present, per buildUpdateInputs's own doc
// comment ("dropping response-only fields like compiled_input").
func TestKbapiTypedInputRoundTrip_allFieldsSurvive(t *testing.T) {
	var in kbapi.PackagePolicyTypedInput
	require.NoError(t, json.Unmarshal([]byte(fullyPopulatedTypedInputJSON), &in))

	// This is update.go's buildUpdateInputs round trip, verbatim.
	b, err := json.Marshal(in)
	require.NoError(t, err)
	var reqIn kbapi.PackagePolicyRequestTypedInput
	require.NoError(t, json.Unmarshal(b, &reqIn))

	// Input-level scalar/pointer fields.
	require.NotNil(t, reqIn.Condition)
	assert.Equal(t, "host.os.family == 'linux'", *reqIn.Condition)
	assert.True(t, reqIn.Enabled)
	require.NotNil(t, reqIn.Id)
	assert.Equal(t, "input-id-1", *reqIn.Id)
	require.NotNil(t, reqIn.KeepEnabled)
	assert.True(t, *reqIn.KeepEnabled)
	require.NotNil(t, reqIn.MigrateFrom)
	assert.Equal(t, "migrate-from-x", *reqIn.MigrateFrom)
	require.NotNil(t, reqIn.Name)
	assert.Equal(t, "input-name-1", *reqIn.Name)
	require.NotNil(t, reqIn.PolicyTemplate)
	assert.Equal(t, "cspm", *reqIn.PolicyTemplate)
	assert.Equal(t, "cloudbeat/cis_aws", reqIn.Type)

	// Input-level Config (anonymous map[string]struct{Frozen,Type,Value}).
	require.NotNil(t, reqIn.Config)
	cfg := *reqIn.Config
	require.Contains(t, cfg, "cfgkey")
	require.NotNil(t, cfg["cfgkey"].Frozen)
	assert.True(t, *cfg["cfgkey"].Frozen)
	require.NotNil(t, cfg["cfgkey"].Type)
	assert.Equal(t, "text", *cfg["cfgkey"].Type)
	assert.Equal(t, "cfgval", cfg["cfgkey"].Value)

	// Input-level Deprecated (named KibanaHTTPAPIsDeprecationInfo type).
	require.NotNil(t, reqIn.Deprecated)
	assert.Equal(t, "dep-desc", reqIn.Deprecated.Description)
	require.NotNil(t, reqIn.Deprecated.Since)
	assert.Equal(t, "9.0.0", *reqIn.Deprecated.Since)
	require.NotNil(t, reqIn.Deprecated.ReplacedBy)
	assert.Equal(t, "new", (*reqIn.Deprecated.ReplacedBy)["old"])

	// Input-level VarGroupSelections and Vars.
	require.NotNil(t, reqIn.VarGroupSelections)
	assert.Equal(t, "opt1", (*reqIn.VarGroupSelections)["vg1"])

	require.NotNil(t, reqIn.Vars)
	vars := *reqIn.Vars
	require.Contains(t, vars, "varkey")
	require.NotNil(t, vars["varkey"].Type)
	assert.Equal(t, "text", *vars["varkey"].Type)
	assert.Equal(t, "varval", vars["varkey"].Value)
	require.NotNil(t, vars["varkey"].Frozen)
	assert.False(t, *vars["varkey"].Frozen)

	// compiled_input is response-only: kbapi.PackagePolicyRequestTypedInput
	// has no field for it, so there is nothing to assert it into -- confirmed
	// here by re-marshaling reqIn and checking the JSON key is genuinely
	// absent, not merely zero-valued.
	reqRaw, err := json.Marshal(reqIn)
	require.NoError(t, err)
	var reqDecoded map[string]any
	require.NoError(t, json.Unmarshal(reqRaw, &reqDecoded))
	_, hasCompiledInput := reqDecoded["compiled_input"]
	assert.False(t, hasCompiledInput, "compiled_input is response-only and must not survive into the request body")

	// Stream-level fields.
	require.NotNil(t, reqIn.Streams)
	streams := *reqIn.Streams
	require.Len(t, streams, 1)
	s := streams[0]

	// Unlike compiled_input, compiled_stream exists on BOTH
	// PackagePolicyTypedInputStream and PackagePolicyRequestTypedInputStream,
	// so it must survive.
	assert.NotNil(t, s.CompiledStream, "compiled_stream exists on both the response and request stream types and must survive")

	require.NotNil(t, s.Condition)
	assert.Equal(t, "data_stream.dataset == 'audit'", *s.Condition)

	require.NotNil(t, s.Config)
	scfg := *s.Config
	require.Contains(t, scfg, "scfgkey")
	require.NotNil(t, scfg["scfgkey"].Frozen)
	assert.True(t, *scfg["scfgkey"].Frozen)
	assert.Equal(t, "scfgval", scfg["scfgkey"].Value)

	assert.Equal(t, "cloud_security_posture.findings", s.DataStream.Dataset)
	require.NotNil(t, s.DataStream.Type)
	assert.Equal(t, "logs", *s.DataStream.Type)
	require.NotNil(t, s.DataStream.Elasticsearch)
	require.NotNil(t, s.DataStream.Elasticsearch.DynamicDataset)
	assert.True(t, *s.DataStream.Elasticsearch.DynamicDataset)
	require.NotNil(t, s.DataStream.Elasticsearch.DynamicNamespace)
	assert.True(t, *s.DataStream.Elasticsearch.DynamicNamespace)
	require.NotNil(t, s.DataStream.Elasticsearch.Privileges)
	require.NotNil(t, s.DataStream.Elasticsearch.Privileges.Indices)
	assert.Equal(t, []string{"idx1", "idx2"}, *s.DataStream.Elasticsearch.Privileges.Indices)

	require.NotNil(t, s.Deprecated)
	assert.Equal(t, "stream-dep-desc", s.Deprecated.Description)

	assert.True(t, s.Enabled)
	require.NotNil(t, s.Id)
	assert.Equal(t, "stream-id-1", *s.Id)
	require.NotNil(t, s.KeepEnabled)
	assert.True(t, *s.KeepEnabled)
	require.NotNil(t, s.MigrateFrom)
	assert.Equal(t, "stream-migrate-from", *s.MigrateFrom)
	require.NotNil(t, s.Release)
	assert.Equal(t, kbapi.PackagePolicyRequestTypedInputStreamRelease("beta"), *s.Release)

	require.NotNil(t, s.VarGroupSelections)
	assert.Equal(t, "sopt1", (*s.VarGroupSelections)["svg1"])

	require.NotNil(t, s.Vars)
	svars := *s.Vars
	require.Contains(t, svars, "streamvarkey")
	require.NotNil(t, svars["streamvarkey"].Frozen)
	assert.True(t, *svars["streamvarkey"].Frozen)
	require.NotNil(t, svars["streamvarkey"].Type)
	assert.Equal(t, "text", *svars["streamvarkey"].Type)
	assert.Equal(t, "streamvarval", svars["streamvarkey"].Value)
}
