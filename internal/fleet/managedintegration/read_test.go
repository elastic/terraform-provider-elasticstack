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

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadAgentlessPolicy_notFound signals removed out-of-band without error.
func TestReadAgentlessPolicy_notFound(t *testing.T) {
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

	out, ok, diags := readAgentlessPolicy(context.Background(), client, "policy-1", "default", seeded)
	require.False(t, diags.HasError(), "%v", diags)
	require.False(t, ok)
	assert.Equal(t, "seed-name-must-not-change", out.Name.ValueString())
}

// TestReadAgentlessPolicy_populatesFromManagedIntegration exercises GET
// /api/fleet/managed_integrations/{id} and populateFromManagedIntegration while
// preserving create-only flags from the incoming model.
func TestReadAgentlessPolicy_populatesFromManagedIntegration(t *testing.T) {
	const managedIntegrationJSON = `{"item":{` +
		`"id":"policy-1","name":"api-name",` +
		`"created_at":"2026-01-01T00:00:00.000Z","created_by":"elastic",` +
		`"updated_at":"2026-01-02T00:00:00.000Z","updated_by":"elastic",` +
		`"inputs":{},` +
		`"package":{"name":"cloud_security_posture","version":"3.4.0"}` +
		`}}`

	mux := http.NewServeMux()
	mux.HandleFunc("/api/fleet/managed_integrations/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, managedIntegrationJSON)
	})
	client := newTopologyTestClient(t, mux)

	seeded := baseTestModel(t)
	seeded.Force = types.BoolValue(true)
	seeded.CreateDatasetTemplates = types.BoolValue(true)
	seeded.SkipTopologyCheck = types.BoolValue(true)
	seeded.ForceDelete = types.BoolValue(true)

	out, ok, diags := readAgentlessPolicy(context.Background(), client, "policy-1", "default", seeded)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, ok)
	assert.Equal(t, "api-name", out.Name.ValueString())
	assert.Equal(t, "default/policy-1", out.ID.ValueString())
	assert.True(t, out.Force.ValueBool())
	assert.True(t, out.CreateDatasetTemplates.ValueBool())
	assert.True(t, out.SkipTopologyCheck.ValueBool())
	assert.True(t, out.ForceDelete.ValueBool())
}
