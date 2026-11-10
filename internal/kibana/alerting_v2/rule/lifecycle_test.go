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

package rule

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

const ruleResponseJSON = `{
	"id": "my-rule",
	"kind": "alert",
	"enabled": true,
	"version": 2,
	"created_at": "2026-01-02T03:04:05.000Z",
	"updated_at": "2026-01-03T03:04:05.000Z",
	"metadata": {"name": "CPU high"},
	"schedule": {"every": "1m"},
	"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
	"recovery": {"strategy": "no_breach"},
	"no_data": {"strategy": "ignore"}
}`

type scriptedResponse struct {
	status int
	body   string
}

type scriptedServer struct {
	responses map[string][]scriptedResponse
	requests  []string
	bodies    map[string][]byte
}

// newScriptedClient serves canned responses keyed by "<METHOD> <path>" (or
// "<METHOD> *" for any path), consuming one per call, and records every request
// so tests can assert which calls were made.
func newScriptedClient(t *testing.T, responses map[string][]scriptedResponse) (*clients.KibanaScopedClient, *scriptedServer) {
	t.Helper()
	script := &scriptedServer{responses: responses, bodies: map[string][]byte{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		script.requests = append(script.requests, key)
		script.bodies[key], _ = io.ReadAll(r.Body)
		queueKey := key
		if _, scripted := script.responses[queueKey]; !scripted {
			queueKey = r.Method + " *"
		}
		queue := script.responses[queueKey]
		if len(queue) == 0 {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		next := queue[0]
		script.responses[queueKey] = queue[1:]
		if next.body != "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(next.status)
		_, _ = w.Write([]byte(next.body))
	}))
	t.Cleanup(server.Close)

	oapi, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: server.URL, Username: "test", Password: "test"})
	require.NoError(t, err)
	return clients.NewKibanaScopedClientForTest(oapi), script
}

func TestReadRule_populatesModelAndIdentity(t *testing.T) {
	t.Parallel()

	client, script := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusOK, body: ruleResponseJSON}},
	})

	model, found, diags := readRule(context.Background(), client, "my-rule", "team-a", Model{})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.True(t, found)
	require.Equal(t, "team-a/my-rule", model.ID.ValueString())
	require.Equal(t, "my-rule", model.RuleID.ValueString())
	require.Equal(t, "team-a", model.SpaceID.ValueString())
	require.Equal(t, "CPU high", model.Metadata.Name.ValueString())
	require.Equal(t, types.BoolValue(true), model.Enabled)
	require.Equal(t, "@timestamp", model.TimeField.ValueString())
	require.Equal(t, []string{"GET /s/team-a/api/alerting/v2/rules/my-rule"}, script.requests)
}

func TestReadRule_notFoundIsNotAnError(t *testing.T) {
	t.Parallel()

	client, _ := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /api/alerting/v2/rules/my-rule": {{status: http.StatusNotFound, body: `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"gone"}`}},
	})

	_, found, diags := readRule(context.Background(), client, "my-rule", "default", Model{})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.False(t, found)
}

func minimalPlan(ruleID string) Model {
	plan := Model{
		Kind:      types.StringValue("alert"),
		Enabled:   types.BoolValue(true),
		TimeField: types.StringValue("@timestamp"),
		Metadata:  &metadataModel{Name: types.StringValue("CPU high")},
		Schedule:  &scheduleModel{Every: types.StringValue("1m")},
		Query:     &queryModel{Base: types.StringValue("FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100")},
		Recovery:  &recoveryModel{Strategy: types.StringValue("no_breach")},
		NoData:    &noDataModel{Strategy: types.StringValue("ignore")},
	}
	if ruleID != "" {
		plan.RuleID = types.StringValue(ruleID)
	} else {
		plan.RuleID = types.StringUnknown()
	}
	return plan
}

func TestCreateRule_refusesToAdoptExistingRule(t *testing.T) {
	t.Parallel()

	client, script := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusOK, body: ruleResponseJSON}},
	})

	_, diags := createRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    minimalPlan("my-rule"),
		SpaceID: "team-a",
		WriteID: "my-rule",
	})

	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "team-a/my-rule")
	require.Contains(t, diags[0].Detail(), "terraform import")
	require.Equal(t, []string{"GET /s/team-a/api/alerting/v2/rules/my-rule"}, script.requests)
}

func TestCreateRule_putsAfterFreeIDCheckAndSetsIdentity(t *testing.T) {
	t.Parallel()

	client, script := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusNotFound, body: `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"gone"}`}},
		"PUT /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusCreated, body: ruleResponseJSON}},
	})

	result, diags := createRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    minimalPlan("my-rule"),
		SpaceID: "team-a",
		WriteID: "my-rule",
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "team-a/my-rule", result.Model.ID.ValueString())
	require.Equal(t, "my-rule", result.Model.RuleID.ValueString())
	require.Equal(t, "team-a", result.Model.SpaceID.ValueString())
	require.Equal(t, []string{
		"GET /s/team-a/api/alerting/v2/rules/my-rule",
		"PUT /s/team-a/api/alerting/v2/rules/my-rule",
	}, script.requests)
	require.JSONEq(t, `{
		"kind": "alert",
		"enabled": true,
		"time_field": "@timestamp",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`, string(script.bodies["PUT /s/team-a/api/alerting/v2/rules/my-rule"]))
}

func TestCreateRule_putAnsweringReplacedIsAnError(t *testing.T) {
	t.Parallel()

	client, _ := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusNotFound, body: `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"gone"}`}},
		"PUT /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusOK, body: ruleResponseJSON}},
	})

	_, diags := createRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    minimalPlan("my-rule"),
		SpaceID: "team-a",
		WriteID: "my-rule",
	})

	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "team-a/my-rule")
	require.Contains(t, diags[0].Detail(), "terraform import")
	require.Contains(t, diags[0].Detail(), "replaced")
}

func TestCreateRule_generatesUUIDWhenRuleIDOmitted(t *testing.T) {
	t.Parallel()

	client, script := newScriptedClient(t, map[string][]scriptedResponse{
		"GET *": {{status: http.StatusNotFound, body: `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"gone"}`}},
		"PUT *": {{status: http.StatusCreated, body: ruleResponseJSON}},
	})

	result, diags := createRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    minimalPlan(""),
		SpaceID: "team-a",
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	generated, err := uuid.Parse(result.Model.RuleID.ValueString())
	require.NoError(t, err)
	require.Equal(t, "team-a/"+generated.String(), result.Model.ID.ValueString())
	require.Equal(t, []string{
		"GET /s/team-a/api/alerting/v2/rules/" + generated.String(),
		"PUT /s/team-a/api/alerting/v2/rules/" + generated.String(),
	}, script.requests)
}

func TestUpdateRule_issuesSinglePutWithFullBody(t *testing.T) {
	t.Parallel()

	client, script := newScriptedClient(t, map[string][]scriptedResponse{
		"PUT /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusOK, body: ruleResponseJSON}},
	})
	plan := minimalPlan("my-rule")
	plan.Enabled = types.BoolValue(false)
	prior := minimalPlan("my-rule")

	result, diags := updateRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    plan,
		Prior:   &prior,
		SpaceID: "team-a",
		WriteID: "my-rule",
	})

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "team-a/my-rule", result.Model.ID.ValueString())
	require.Equal(t, "my-rule", result.Model.RuleID.ValueString())
	require.Equal(t, []string{"PUT /s/team-a/api/alerting/v2/rules/my-rule"}, script.requests)
	require.JSONEq(t, `{
		"kind": "alert",
		"enabled": false,
		"time_field": "@timestamp",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`, string(script.bodies["PUT /s/team-a/api/alerting/v2/rules/my-rule"]))
}

func TestUpdateRule_putAnsweringCreatedIsAnError(t *testing.T) {
	t.Parallel()

	client, _ := newScriptedClient(t, map[string][]scriptedResponse{
		"PUT /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusCreated, body: ruleResponseJSON}},
	})
	prior := minimalPlan("my-rule")

	_, diags := updateRule(context.Background(), client, entitycore.KibanaWriteRequest[Model]{
		Plan:    minimalPlan("my-rule"),
		Prior:   &prior,
		SpaceID: "team-a",
		WriteID: "my-rule",
	})

	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "team-a/my-rule")
	require.Contains(t, diags[0].Detail(), "deleted outside Terraform")
	require.Contains(t, diags[0].Detail(), "recreated")
	require.Contains(t, diags[0].Detail(), "plan")
}

func TestReadRule_errorDiagnosticsCarryCompositeID(t *testing.T) {
	t.Parallel()

	client, _ := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusInternalServerError, body: `{"code":"INTERNAL","error":"Internal Server Error","message":"boom"}`}},
	})

	_, found, diags := readRule(context.Background(), client, "my-rule", "team-a", Model{})

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "team-a/my-rule")
	require.Contains(t, diags[0].Detail(), "boom")
}

func TestWithCompositeID_prefixesErrorsAndKeepsWarnings(t *testing.T) {
	t.Parallel()

	in := diag.Diagnostics{
		diag.NewWarningDiagnostic("heads up", "builder present"),
		diag.NewErrorDiagnostic("failed", "boom"),
	}

	out := withCompositeID(in, "team-a", "my-rule")

	require.Len(t, out, 2)
	require.Equal(t, in[0], out[0])
	require.Equal(t, "failed", out[1].Summary())
	require.Equal(t, `Rule "team-a/my-rule": boom`, out[1].Detail())
}

func TestReadRule_mappingErrorDiagnosticsCarryCompositeID(t *testing.T) {
	t.Parallel()

	unknownStrategy := strings.Replace(ruleResponseJSON, `"recovery": {"strategy": "no_breach"}`, `"recovery": {"strategy": "teleport"}`, 1)
	client, _ := newScriptedClient(t, map[string][]scriptedResponse{
		"GET /s/team-a/api/alerting/v2/rules/my-rule": {{status: http.StatusOK, body: unknownStrategy}},
	})

	_, found, diags := readRule(context.Background(), client, "my-rule", "team-a", Model{})

	require.False(t, found)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "team-a/my-rule")
	require.Contains(t, diags[0].Detail(), "teleport")
}
