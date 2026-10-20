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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func accessControlObject(t *testing.T, mode string) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(accessControlAttrTypes(), map[string]attr.Value{
		"access_mode": types.StringValue(mode),
	})
	require.False(t, diags.HasError(), "%v", diags)
	return obj
}

func nullAccessControl() types.Object { return types.ObjectNull(accessControlAttrTypes()) }

func baseAgentModel() agentModel {
	m := agentModel{}
	m.Name = types.StringValue("n")
	m.AgentID = types.StringValue("a")
	m.Tools = types.SetNull(types.StringType)
	m.Labels = types.SetNull(types.StringType)
	m.SkillIDs = types.SetNull(types.StringType)
	return m
}

func marshal(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func TestToAPICreateModel_accessControlSet(t *testing.T) {
	m := baseAgentModel()
	m.AccessControl = accessControlObject(t, "private")
	body, diags := m.toAPICreateModel(context.Background(), true)
	require.False(t, diags.HasError())
	require.Equal(t, map[string]any{"access_mode": "private"}, marshal(t, body)["access_control"])
}

func TestToAPICreateModel_accessControlOmittedWhenNullOrUnknown(t *testing.T) {
	for name, v := range map[string]types.Object{
		"null":    nullAccessControl(),
		"unknown": types.ObjectUnknown(accessControlAttrTypes()),
	} {
		t.Run(name, func(t *testing.T) {
			m := baseAgentModel()
			m.AccessControl = v
			body, diags := m.toAPICreateModel(context.Background(), true)
			require.False(t, diags.HasError())
			require.NotContains(t, marshal(t, body), "access_control")
		})
	}
}

func TestToAPIUpdateModel_accessControl(t *testing.T) {
	m := baseAgentModel()
	m.AccessControl = accessControlObject(t, "shared")

	body, diags := m.toAPIUpdateModel(context.Background(), true, true)
	require.False(t, diags.HasError())
	require.Equal(t, map[string]any{"access_mode": "shared"}, marshal(t, body)["access_control"])

	body, diags = m.toAPIUpdateModel(context.Background(), true, false)
	require.False(t, diags.HasError())
	require.NotContains(t, marshal(t, body), "access_control")
}

func TestConfigSendsAccessControl(t *testing.T) {
	require.True(t, sendsAccessControl(accessControlObject(t, "public")))
	require.False(t, sendsAccessControl(nullAccessControl()))
	require.False(t, sendsAccessControl(types.ObjectUnknown(accessControlAttrTypes())))
}

func TestPopulateAccessControl(t *testing.T) {
	m := baseAgentModel()
	m.AccessControl = accessControlObject(t, "shared")

	diags := m.populateAccessControl(&models.AgentAccessControl{AccessMode: "private"})
	require.False(t, diags.HasError())
	require.Equal(t, accessControlObject(t, "private"), m.AccessControl)

	diags = m.populateAccessControl(nil)
	require.False(t, diags.HasError())
	require.True(t, m.AccessControl.IsNull())
}

func TestAgentAccessControlResponse_decode(t *testing.T) {
	var r models.AgentAccessControlResponse
	require.NoError(t, json.Unmarshal([]byte(`{"access_control":{"access_mode":"shared","entries":[]},"permissions":{}}`), &r))
	require.Equal(t, "shared", r.AccessControl.AccessMode)
}
