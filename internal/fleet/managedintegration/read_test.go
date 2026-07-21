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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadAgentlessPolicy_typedPackagePolicyInputsPreservesModelOnBridgeError
// proves readAgentlessPolicy returns bridge diagnostics and leaves the incoming
// model untouched when GET /api/fleet/package_policies/{id} returns a typed
// (non-mapped) inputs array — the shape the compat client still uses until
// task 8.
func TestReadAgentlessPolicy_typedPackagePolicyInputsPreservesModelOnBridgeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/package_policies/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"item":%s}`, typedArrayPackagePolicyForBridgeJSON)
	})
	client := newTopologyTestClient(t, mux)

	seeded := baseTestModel(t)
	seeded.PolicyID = types.StringValue("policy-1")
	seeded.ID = types.StringValue("default/policy-1")
	seeded.Name = types.StringValue("seed-name-must-not-change")
	seeded.Description = types.StringValue("seed-description")

	out, ok, diags := readAgentlessPolicy(context.Background(), client, "policy-1", "default", seeded)
	require.True(t, diags.HasError(), "%v", diags)
	require.False(t, ok)
	requireDiagnosticAtPath(t, diags, path.Root("inputs"), "Unexpected package policy inputs format")

	assert.Equal(t, "seed-name-must-not-change", out.Name.ValueString())
	assert.Equal(t, "seed-description", out.Description.ValueString())
	assert.Equal(t, "default/policy-1", out.ID.ValueString())
}
