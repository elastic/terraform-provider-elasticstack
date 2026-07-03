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

package policyshape

import "encoding/json"

// TypedVarEntry mirrors the `{frozen, type, value}` shape used for `vars`
// (and, in the elastic_defend_integration_policy resource, `config`)
// throughout the "typed" package-policy request/response family
// (KibanaHTTPAPIsUpdatePackagePolicyRequest, PackagePolicyRequestTypedInput,
// PackagePolicyRequestTypedInputStream, PackagePolicyTypedInput, and
// PackagePolicyTypedInputStream). oapi-codegen gives each occurrence its own
// anonymous map[string]struct{Frozen,Type,Value} Go type (structurally
// identical, but not always assignable/convertible to each other -- see e.g.
// PackagePolicyTypedInputStream.Release vs
// PackagePolicyRequestTypedInputStream.Release, which alias the same
// underlying string type via two distinct named types).
//
// This is declared as a type ALIAS (not a defined type) rather than a
// `type TypedVarEntry struct {...}` declaration: an alias is structurally
// identical to -- and therefore directly assignable to/from -- every one of
// those anonymous generated field types (e.g.
// `input.Config = &map[string]TypedVarEntry{...}` compiles without a
// conversion), whereas a defined type would not be. Callers that need to
// merge across several different anonymous field types generically (see
// agentlesspolicy/update.go's mergeVarsInto) can still do so via a JSON
// marshal/unmarshal round trip instead of hand-spelling each anonymous type.
type TypedVarEntry = struct {
	Frozen *bool   `json:"frozen,omitempty"`
	Type   *string `json:"type,omitempty"`
	Value  any     `json:"value,omitempty"`
}

// The Fleet OpenAPI spec models each `vars` value as a per-value union
// (string | number | bool | []string | []float | {id,isSecretRef}). oapi-codegen
// generates a distinct wrapper struct per endpoint
// (e.g. KibanaHTTPAPIsPackagePolicyResponse_Vars,
// PackagePolicyMappedInput_Vars_AdditionalProperties, etc.) all with the same
// JSON representation. Callers treat vars as opaque JSON (vars_json), so the
// helpers below convert between any of those generated wrappers and a flat
// map[string]any via JSON round-trip. Marshalling a typed map of unions yields
// {var_name: <raw_value>}; unmarshalling a flat map back into the typed map
// dispatches each value through the union's UnmarshalJSON.

// VarsAnyToMap marshals any vars-shaped value (a *struct wrapper or a typed
// map of wrappers) to a flat map[string]any. Returns nil when the input is
// nil, empty, or serialises to JSON null.
func VarsAnyToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	bytes, err := json.Marshal(v)
	if err != nil || len(bytes) == 0 || string(bytes) == "null" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil
	}
	return out
}

// VarsMapToTypedMap converts a flat map[string]any to a typed
// map[string]*T pointer expected by request bodies. Returns nil when m is
// empty so the wire payload omits the `vars` field entirely.
func VarsMapToTypedMap[T any](m map[string]any) *map[string]*T {
	if len(m) == 0 {
		return nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var out map[string]*T
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil
	}
	return &out
}

// VarsMapToUnionWrapper packs a flat map[string]any into a struct wrapper
// (e.g. KibanaHTTPAPIsPackagePolicyResponse_Vars) by JSON round-trip. Returns
// nil when m is empty.
func VarsMapToUnionWrapper[T any](m map[string]any) *T {
	if len(m) == 0 {
		return nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var out T
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil
	}
	return &out
}
