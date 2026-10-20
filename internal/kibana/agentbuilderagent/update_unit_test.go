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

package agentbuilderagent

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

// runUpdateAgent drives updateAgent against a fake Kibana and returns the
// decoded PUT request body.
func runUpdateAgent(t *testing.T, plan, config agentModel) map[string]any {
	t.Helper()
	var putBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/status":
			_, _ = w.Write([]byte(`{"version":{"number":"9.5.0","build_flavor":"traditional"}}`))
		case r.Method == http.MethodPut:
			putBody, _ = io.ReadAll(r.Body)
			_, _ = w.Write([]byte(`{"id":"a","name":"n"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	oapi, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: srv.URL})
	require.NoError(t, err)
	client := clients.NewKibanaScopedClientForTest(oapi)

	_, diags := updateAgent(context.Background(), client, entitycore.KibanaWriteRequest[agentModel]{
		Plan:    plan,
		Config:  config,
		WriteID: "a",
		SpaceID: "default",
	})
	require.False(t, diags.HasError(), "%v", diags)
	require.NotNil(t, putBody, "no PUT request captured")
	return marshalBytes(t, putBody)
}

func marshalBytes(t *testing.T, b []byte) map[string]any {
	t.Helper()
	return marshal(t, jsonRaw(b))
}

type jsonRaw []byte

func (j jsonRaw) MarshalJSON() ([]byte, error) { return j, nil }

func TestUpdateAgent_omitsAccessControlWhenConfigNull(t *testing.T) {
	plan := baseAgentModel()
	plan.AccessControl = accessControlObject(t, "shared") // carried forward by UseStateForUnknown
	config := baseAgentModel()
	config.AccessControl = nullAccessControl()

	body := runUpdateAgent(t, plan, config)
	require.NotContains(t, body, "access_control")
	require.Equal(t, "n", body["name"])
}

func TestUpdateAgent_sendsAccessControlWhenConfigured(t *testing.T) {
	plan := baseAgentModel()
	plan.AccessControl = accessControlObject(t, "shared")
	config := baseAgentModel()
	config.AccessControl = accessControlObject(t, "shared")

	body := runUpdateAgent(t, plan, config)
	require.Equal(t, map[string]any{"access_mode": "shared"}, body["access_control"])
}

func TestUpdateAgent_avatarOmittedWhenConfigNull(t *testing.T) {
	plan := baseAgentModel()
	plan.AvatarColor = types.StringValue("#BFDBFF") // carried forward by UseStateForUnknown
	plan.AvatarSymbol = types.StringValue("TA")
	config := baseAgentModel()
	config.AccessControl = nullAccessControl()
	config.AvatarColor = types.StringNull()
	config.AvatarSymbol = types.StringNull()

	body := runUpdateAgent(t, plan, config)
	require.NotContains(t, body, "avatar_color")
	require.NotContains(t, body, "avatar_symbol")
}

func TestUpdateAgent_avatarSentWhenConfigured(t *testing.T) {
	plan := baseAgentModel()
	plan.AvatarColor = types.StringValue("")
	plan.AvatarSymbol = types.StringValue("TA")
	config := baseAgentModel()
	config.AccessControl = nullAccessControl()
	config.AvatarColor = types.StringValue("") // explicit clear
	config.AvatarSymbol = types.StringValue("TA")

	body := runUpdateAgent(t, plan, config)
	require.Equal(t, "", body["avatar_color"])
	require.Equal(t, "TA", body["avatar_symbol"])
}
