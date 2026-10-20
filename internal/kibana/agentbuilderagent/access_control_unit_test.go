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

	"errors"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	return agentModel{
		agentBaseModel: agentBaseModel{ //nolint:modernize // promoted fields cannot be set in a composite literal
			AgentID:  types.StringValue("a"),
			Name:     types.StringValue("n"),
			Labels:   types.SetNull(types.StringType),
			SkillIDs: types.SetNull(types.StringType),
		},
		Tools: types.SetNull(types.StringType),
	}
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

	diags := m.populateAccessControl(context.Background(), &models.AgentAccessControl{AccessMode: "private"})
	require.False(t, diags.HasError())
	require.Equal(t, accessControlObject(t, "private"), m.AccessControl)

	diags = m.populateAccessControl(context.Background(), nil)
	require.False(t, diags.HasError())
	require.True(t, m.AccessControl.IsNull())
}

func TestAgentAccessControlResponse_decode(t *testing.T) {
	var r models.AgentAccessControlResponse
	require.NoError(t, json.Unmarshal([]byte(`{"access_control":{"access_mode":"shared","entries":[]},"permissions":{}}`), &r))
	require.Equal(t, "shared", r.AccessControl.AccessMode)
}

type readCalls struct{ accessControl int }

func testReader(agent *models.Agent, supported bool, ac *models.AgentAccessControl, acDiags diag.Diagnostics, calls *readCalls) agentReader {
	return agentReader{
		getAgent: func(context.Context) (*models.Agent, diag.Diagnostics) { return agent, nil },
		supportsAccessControl: func(context.Context) (bool, diag.Diagnostics) {
			return supported, nil
		},
		getAgentAccessControl: func(context.Context) (*models.AgentAccessControl, diag.Diagnostics) {
			calls.accessControl++
			return ac, acDiags
		},
	}
}

func TestReadAgentWith_olderStackClearsCarriedForwardMode(t *testing.T) {
	calls := &readCalls{}
	prior := baseAgentModel()
	prior.AccessControl = accessControlObject(t, "shared")

	got, found, diags := readAgentWith(context.Background(), testReader(&models.Agent{ID: "a", Name: "n"}, false, nil, nil, calls), "default", prior)
	require.False(t, diags.HasError())
	require.True(t, found)
	require.True(t, got.AccessControl.IsNull())
	require.Zero(t, calls.accessControl)
}

func TestReadAgentWith_missingAgentSkipsAccessControl(t *testing.T) {
	calls := &readCalls{}
	_, found, diags := readAgentWith(context.Background(), testReader(nil, true, nil, nil, calls), "default", baseAgentModel())
	require.False(t, diags.HasError())
	require.False(t, found)
	require.Zero(t, calls.accessControl)
}

func TestReadAgentWith_supportedStackPopulatesMode(t *testing.T) {
	calls := &readCalls{}
	reader := testReader(&models.Agent{ID: "a", Name: "n"}, true, &models.AgentAccessControl{AccessMode: "private"}, nil, calls)
	got, found, diags := readAgentWith(context.Background(), reader, "default", baseAgentModel())
	require.False(t, diags.HasError())
	require.True(t, found)
	require.Equal(t, accessControlObject(t, "private"), got.AccessControl)
	require.Equal(t, 1, calls.accessControl)
}

func TestReadAgentWith_accessControlFailuresAreErrors(t *testing.T) {
	forbidden := diag.Diagnostics{diag.NewErrorDiagnostic("forbidden", errors.New("403").Error())}
	for name, tc := range map[string]struct {
		ac    *models.AgentAccessControl
		diags diag.Diagnostics
	}{
		"403":                {diags: forbidden},
		"404 after main GET": {},
	} {
		t.Run(name, func(t *testing.T) {
			calls := &readCalls{}
			_, found, diags := readAgentWith(context.Background(), testReader(&models.Agent{ID: "a", Name: "n"}, true, tc.ac, tc.diags, calls), "default", baseAgentModel())
			require.True(t, diags.HasError())
			require.False(t, found)
		})
	}
}
