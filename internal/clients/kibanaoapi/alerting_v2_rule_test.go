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

package kibanaoapi_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/require"
)

const alertingV2RuleJSON = `{
	"id": "my-rule",
	"kind": "alert",
	"enabled": true,
	"version": 3,
	"created_at": "2026-01-02T03:04:05.000Z",
	"updated_at": "2026-01-03T03:04:05.000Z",
	"metadata": {"name": "CPU high"},
	"schedule": {"every": "1m"},
	"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
	"recovery": {"strategy": "no_breach"},
	"no_data": {"strategy": "ignore"}
}`

type recordedRequest struct {
	method string
	path   string
	body   []byte
}

func newAlertingV2Client(t *testing.T, statusCode int, responseBody string, recorded *recordedRequest) *kibanaoapi.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if recorded != nil {
			recorded.method = r.Method
			recorded.path = r.URL.Path
			recorded.body, _ = io.ReadAll(r.Body)
		}
		if responseBody != "" {
			w.Header().Set("Content-Type", "application/json")
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(server.Close)

	client, err := kibanaoapi.NewClient(kibanaoapi.Config{URL: server.URL, Username: "test", Password: "test"})
	require.NoError(t, err)
	return client
}

func TestGetAlertingV2Rule_returnsRuleFromSpacePrefixedPath(t *testing.T) {
	t.Parallel()

	var recorded recordedRequest
	client := newAlertingV2Client(t, http.StatusOK, alertingV2RuleJSON, &recorded)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "team-a", "my-rule")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.NotNil(t, rule)
	require.Equal(t, "my-rule", rule.Id)
	require.Equal(t, 3, rule.Version)
	require.Equal(t, http.MethodGet, recorded.method)
	require.Equal(t, "/s/team-a/api/alerting/v2/rules/my-rule", recorded.path)
}

func TestGetAlertingV2Rule_notFoundReturnsNilWithoutError(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusNotFound, `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"Rule my-rule not found"}`, nil)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Nil(t, rule)
}

func TestGetAlertingV2Rule_alertingDisabledNamesAdvancedSetting(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusServiceUnavailable, `{"code":"ALERTING_DISABLED","error":"Service Unavailable","message":"Alerting v2 is disabled"}`, nil)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.Nil(t, rule)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "alerting:v2:enabled")
	require.Contains(t, diags[0].Detail(), "Alerting v2 is disabled")
}

func TestGetAlertingV2Rule_errorBodySurfacesCodeAndMessage(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusBadRequest, `{"code":"INVALID_SCHEDULE","error":"Bad Request","message":"schedule.every is below the minimum"}`, nil)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.Nil(t, rule)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "INVALID_SCHEDULE")
	require.Contains(t, diags[0].Summary(), "400")
	require.Equal(t, "schedule.every is below the minimum", diags[0].Detail())
}

func TestGetAlertingV2Rule_emptySuccessBodyIsError(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusOK, "", nil)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.Nil(t, rule)
	require.True(t, diags.HasError())
}

func TestGetAlertingV2Rule_defaultSpaceUsesUnprefixedPath(t *testing.T) {
	t.Parallel()

	var recorded recordedRequest
	client := newAlertingV2Client(t, http.StatusOK, alertingV2RuleJSON, &recorded)

	_, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "/api/alerting/v2/rules/my-rule", recorded.path)
}

func minimalPutRuleBody() kbapi.KibanaHTTPAPIsAlertingPutRule {
	enabled := true
	body := kbapi.KibanaHTTPAPIsAlertingPutRule{
		Kind:     kbapi.KibanaHTTPAPIsAlertingPutRuleKindAlert,
		Enabled:  &enabled,
		Metadata: kbapi.KibanaHTTPAPIsAlertingRuleMetadata{Name: "CPU high"},
		Schedule: kbapi.KibanaHTTPAPIsAlertingRuleSchedule{Every: "1m"},
		Query:    kbapi.KibanaHTTPAPIsAlertingRuleQuery{Base: "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
	}
	return body
}

func TestPutAlertingV2Rule_distinguishesCreatedFromReplaced(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		status  int
		created bool
	}{
		"201 means created":  {status: http.StatusCreated, created: true},
		"200 means replaced": {status: http.StatusOK, created: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var recorded recordedRequest
			client := newAlertingV2Client(t, tc.status, alertingV2RuleJSON, &recorded)

			created, diags := kibanaoapi.PutAlertingV2Rule(context.Background(), client, "team-a", "my-rule", minimalPutRuleBody())

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			require.Equal(t, tc.created, created)
			require.Equal(t, http.MethodPut, recorded.method)
			require.Equal(t, "/s/team-a/api/alerting/v2/rules/my-rule", recorded.path)
			require.JSONEq(t, `{
				"kind": "alert",
				"enabled": true,
				"metadata": {"name": "CPU high"},
				"schedule": {"every": "1m"},
				"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"}
			}`, string(recorded.body))
		})
	}
}

func TestPutAlertingV2Rule_conflictSurfacesCodeAndMessage(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusConflict, `{"code":"RULE_NAME_CONFLICT","error":"Conflict","message":"A rule named CPU high already exists"}`, nil)

	created, diags := kibanaoapi.PutAlertingV2Rule(context.Background(), client, "default", "my-rule", minimalPutRuleBody())

	require.False(t, created)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "RULE_NAME_CONFLICT")
	require.Equal(t, "A rule named CPU high already exists", diags[0].Detail())
}

func TestPutAlertingV2Rule_alertingDisabledNamesAdvancedSetting(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusServiceUnavailable, `{"code":"ALERTING_DISABLED","error":"Service Unavailable","message":"Alerting v2 is disabled"}`, nil)

	created, diags := kibanaoapi.PutAlertingV2Rule(context.Background(), client, "default", "my-rule", minimalPutRuleBody())

	require.False(t, created)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "alerting:v2:enabled")
	require.Contains(t, diags[0].Detail(), "Alerting v2 is disabled")
}

func TestPutAlertingV2Rule_emptySuccessBodyIsError(t *testing.T) {
	t.Parallel()

	for name, status := range map[string]int{"201": http.StatusCreated, "200": http.StatusOK} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			client := newAlertingV2Client(t, status, "", nil)

			_, diags := kibanaoapi.PutAlertingV2Rule(context.Background(), client, "default", "my-rule", minimalPutRuleBody())

			require.True(t, diags.HasError())
		})
	}
}

func TestDeleteAlertingV2Rule_statusMapping(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		status      int
		body        string
		wantError   bool
		wantInError string
	}{
		"204 no body is success": {status: http.StatusNoContent},
		"404 is success":         {status: http.StatusNotFound, body: `{"code":"RULE_NOT_FOUND","error":"Not Found","message":"gone"}`},
		"503 alerting disabled": {
			status:      http.StatusServiceUnavailable,
			body:        `{"code":"ALERTING_DISABLED","error":"Service Unavailable","message":"disabled"}`,
			wantError:   true,
			wantInError: "alerting:v2:enabled",
		},
		"500 surfaces message": {status: http.StatusInternalServerError, body: `{"code":"INTERNAL","error":"Internal Server Error","message":"boom"}`, wantError: true, wantInError: "boom"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var recorded recordedRequest
			client := newAlertingV2Client(t, tc.status, tc.body, &recorded)

			diags := kibanaoapi.DeleteAlertingV2Rule(context.Background(), client, "team-a", "my-rule")

			require.Equal(t, http.MethodDelete, recorded.method)
			require.Equal(t, "/s/team-a/api/alerting/v2/rules/my-rule", recorded.path)
			if !tc.wantError {
				require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
				return
			}
			require.True(t, diags.HasError())
			require.Contains(t, diags[0].Detail(), tc.wantInError)
		})
	}
}

func TestGetAlertingV2Rule_nonJSONErrorBodyFallsBackToRawBody(t *testing.T) {
	t.Parallel()

	client := newAlertingV2Client(t, http.StatusBadGateway, "upstream unavailable", nil)

	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, "default", "my-rule")

	require.Nil(t, rule)
	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Summary(), "502")
	require.Equal(t, "upstream unavailable", diags[0].Detail())
}
