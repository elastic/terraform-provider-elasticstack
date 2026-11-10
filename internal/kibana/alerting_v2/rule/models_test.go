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
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestModel_versionRequirementsGateAt960(t *testing.T) {
	t.Parallel()

	reqs, diags := Model{}.GetVersionRequirements(context.Background())

	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.Equal(t, "9.6.0-SNAPSHOT", reqs[0].MinVersion.String())
	require.Contains(t, reqs[0].ErrorMessage, "Alerting v2 rules require Elastic Stack v9.6.0 or later")
}

func stringList(values ...string) types.List {
	elements := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elements = append(elements, types.StringValue(value))
	}
	return types.ListValueMust(types.StringType, elements)
}

func TestModel_toAPIPutBody_mapsEveryConfiguredField(t *testing.T) {
	t.Parallel()

	model := Model{
		Kind:      types.StringValue("alert"),
		Enabled:   types.BoolValue(false),
		TimeField: types.StringValue("event.ingested"),
		Metadata: &metadataModel{
			Name:        types.StringValue("CPU high"),
			Description: types.StringValue("Fires when hosts are busy"),
			Tags:        stringList("z", "a"),
			RoutingTags: stringList("pager"),
		},
		Schedule: &scheduleModel{Every: types.StringValue("1m"), Lookback: types.StringValue("5m")},
		Query: &queryModel{
			Base:   types.StringValue("FROM logs-* | STATS count = COUNT(*) BY host.name"),
			Breach: &breachModel{Segment: types.StringValue("WHERE count > 500")},
		},
		Recovery: &recoveryModel{Strategy: types.StringValue("condition"), Segment: types.StringValue("WHERE count < 50")},
		NoData:   &noDataModel{Strategy: types.StringValue("keep_last"), Query: types.StringValue("FROM logs-* | STATS count = COUNT(*)")},
		StateTransition: &stateTransitionModel{
			Pending:    &stateTransitionPhaseModel{Count: types.Int64Value(2), Timeframe: types.StringValue("5m"), Operator: types.StringValue("and")},
			Recovering: &stateTransitionPhaseModel{Count: types.Int64Value(1)},
		},
		Grouping: &groupingModel{Fields: stringList("host.name", "service.name")},
		Artifacts: []artifactModel{
			{ID: types.StringValue("rb"), Type: types.StringValue("runbook"), Data: jsontypes.NewNormalizedValue(`{"content": "# Steps"}`)},
			{ID: types.StringValue("db"), Type: types.StringValue("dashboard"), Data: jsontypes.NewNormalizedValue(`{"dashboard_id": "abc", "extra": {"n": 1}}`)},
		},
	}

	body, diags := model.toAPIPutBody(context.Background())
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"kind": "alert",
		"enabled": false,
		"time_field": "event.ingested",
		"metadata": {
			"name": "CPU high",
			"description": "Fires when hosts are busy",
			"tags": ["z", "a"],
			"routing_tags": ["pager"]
		},
		"schedule": {"every": "1m", "lookback": "5m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name", "breach": {"segment": "WHERE count > 500"}},
		"recovery": {"strategy": "condition", "segment": "WHERE count < 50"},
		"no_data": {"strategy": "keep_last", "query": "FROM logs-* | STATS count = COUNT(*)"},
		"state_transition": {
			"pending": {"count": 2, "timeframe": "5m", "operator": "and"},
			"recovering": {"count": 1}
		},
		"grouping": {"fields": ["host.name", "service.name"]},
		"artifacts": [
			{"id": "rb", "type": "runbook", "data": {"content": "# Steps"}},
			{"id": "db", "type": "dashboard", "data": {"dashboard_id": "abc", "extra": {"n": 1}}}
		]
	}`, string(raw))
}

func apiRuleResponse(t *testing.T, raw string) *kbapi.KibanaHTTPAPIsAlertingRuleResponse {
	t.Helper()
	var resp kbapi.KibanaHTTPAPIsAlertingRuleResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))
	return &resp
}

func TestModel_populateFromAPI_mapsResponseAndDefaultsTimeField(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": false,
		"version": 4,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-03T06:07:08.000Z",
		"metadata": {"name": "CPU high", "builder": {"type": "threshold"}},
		"schedule": {"every": "1m", "lookback": "5m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
		"recovery": {"strategy": "query", "query": "FROM logs-* | WHERE ok == true"},
		"no_data": {"strategy": "resolve"}
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), resp)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	require.Equal(t, "alert", model.Kind.ValueString())
	require.False(t, model.Enabled.ValueBool())
	require.Equal(t, "@timestamp", model.TimeField.ValueString())
	require.Equal(t, "CPU high", model.Metadata.Name.ValueString())
	require.Equal(t, "1m", model.Schedule.Every.ValueString())
	require.Equal(t, "5m", model.Schedule.Lookback.ValueString())
	require.Equal(t, "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100", model.Query.Base.ValueString())
	require.Equal(t, "query", model.Recovery.Strategy.ValueString())
	require.Equal(t, "FROM logs-* | WHERE ok == true", model.Recovery.Query.ValueString())
	require.True(t, model.Recovery.Segment.IsNull())
	require.Equal(t, "resolve", model.NoData.Strategy.ValueString())
	require.True(t, model.NoData.Query.IsNull())
	require.Equal(t, int64(4), model.Version.ValueInt64())
	require.Equal(t, "2026-01-02T03:04:05.000Z", model.CreatedAt.ValueString())
	require.Equal(t, "2026-01-03T06:07:08.000Z", model.UpdatedAt.ValueString())
}

func TestModel_populateFromAPI_unknownRecoveryStrategyIsError(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "teleport"},
		"no_data": {"strategy": "ignore"}
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), resp)

	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "teleport")
}

func TestModel_unionVariantsRoundTrip(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		recovery     *recoveryModel
		noData       *noDataModel
		wantRecovery string
		wantNoData   string
	}{
		"no_breach and ignore": {
			recovery:     &recoveryModel{Strategy: types.StringValue("no_breach")},
			noData:       &noDataModel{Strategy: types.StringValue("ignore")},
			wantRecovery: `{"strategy":"no_breach"}`,
			wantNoData:   `{"strategy":"ignore"}`,
		},
		"manual and resolve with query": {
			recovery:     &recoveryModel{Strategy: types.StringValue("manual")},
			noData:       &noDataModel{Strategy: types.StringValue("resolve"), Query: types.StringValue("FROM logs-* | STATS c = COUNT(*)")},
			wantRecovery: `{"strategy":"manual"}`,
			wantNoData:   `{"strategy":"resolve","query":"FROM logs-* | STATS c = COUNT(*)"}`,
		},
		"condition and keep_last with query": {
			recovery:     &recoveryModel{Strategy: types.StringValue("condition"), Segment: types.StringValue("WHERE count < 50")},
			noData:       &noDataModel{Strategy: types.StringValue("keep_last"), Query: types.StringValue("FROM logs-* | STATS c = COUNT()")},
			wantRecovery: `{"strategy":"condition","segment":"WHERE count < 50"}`,
			wantNoData:   `{"strategy":"keep_last","query":"FROM logs-* | STATS c = COUNT()"}`,
		},
		"query and keep_last without query": {
			recovery:     &recoveryModel{Strategy: types.StringValue("query"), Query: types.StringValue("FROM logs-* | WHERE ok == true")},
			noData:       &noDataModel{Strategy: types.StringValue("keep_last")},
			wantRecovery: `{"strategy":"query","query":"FROM logs-* | WHERE ok == true"}`,
			wantNoData:   `{"strategy":"keep_last"}`,
		},
		"signal rule has neither": {},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			model := Model{
				Kind:     types.StringValue("signal"),
				Enabled:  types.BoolValue(true),
				Metadata: &metadataModel{Name: types.StringValue("n")},
				Schedule: &scheduleModel{Every: types.StringValue("1m")},
				Query:    &queryModel{Base: types.StringValue("FROM logs-*")},
				Recovery: tc.recovery,
				NoData:   tc.noData,
			}

			body, diags := model.toAPIPutBody(context.Background())
			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

			if tc.recovery == nil {
				require.Nil(t, body.Recovery)
				require.Nil(t, body.NoData)
			} else {
				recoveryJSON, err := json.Marshal(body.Recovery)
				require.NoError(t, err)
				require.JSONEq(t, tc.wantRecovery, string(recoveryJSON))
				noDataJSON, err := json.Marshal(body.NoData)
				require.NoError(t, err)
				require.JSONEq(t, tc.wantNoData, string(noDataJSON))
			}

			readBack := &kbapi.KibanaHTTPAPIsAlertingRuleResponse{Recovery: body.Recovery, NoData: body.NoData}
			var roundTripped Model
			require.False(t, roundTripped.populateFromAPI(context.Background(), readBack).HasError())
			require.Equal(t, tc.recovery, roundTripped.Recovery)
			require.Equal(t, tc.noData, roundTripped.NoData)
		})
	}
}

func TestModel_populateFromAPI_mapsEveryNestedField(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high", "description": "Busy hosts", "tags": ["z", "a"], "routing_tags": ["pager"]},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*", "breach": {"segment": "WHERE count > 500"}},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"},
		"state_transition": {"pending": {"count": 2, "timeframe": "5m", "operator": "and"}, "recovering": {"count": 1}},
		"grouping": {"fields": ["host.name", "service.name"]},
		"artifacts": [
			{"id": "rb", "type": "runbook", "data": {"content": "# Steps"}},
			{"id": "db", "type": "dashboard", "data": {"dashboard_id": "abc", "extra": {"n": 1}}}
		]
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), resp)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	require.Equal(t, "Busy hosts", model.Metadata.Description.ValueString())
	require.Equal(t, stringList("z", "a"), model.Metadata.Tags)
	require.Equal(t, stringList("pager"), model.Metadata.RoutingTags)
	require.Equal(t, "WHERE count > 500", model.Query.Breach.Segment.ValueString())
	require.Equal(t, &stateTransitionPhaseModel{Count: types.Int64Value(2), Timeframe: types.StringValue("5m"), Operator: types.StringValue("and")}, model.StateTransition.Pending)
	require.Equal(t, &stateTransitionPhaseModel{Count: types.Int64Value(1), Timeframe: types.StringNull(), Operator: types.StringNull()}, model.StateTransition.Recovering)
	require.Equal(t, stringList("host.name", "service.name"), model.Grouping.Fields)
	require.Len(t, model.Artifacts, 2)
	require.Equal(t, "rb", model.Artifacts[0].ID.ValueString())
	require.Equal(t, "runbook", model.Artifacts[0].Type.ValueString())
	require.JSONEq(t, `{"content": "# Steps"}`, model.Artifacts[0].Data.ValueString())
	require.Equal(t, "db", model.Artifacts[1].ID.ValueString())
	require.JSONEq(t, `{"dashboard_id": "abc", "extra": {"n": 1}}`, model.Artifacts[1].Data.ValueString())
}

func TestModel_optionalObjectsAndCollections_absentIsNullAndNullIsAbsent(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "signal",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"}
	}`)

	var model Model
	require.False(t, model.populateFromAPI(context.Background(), resp).HasError())

	require.True(t, model.Metadata.Description.IsNull())
	require.True(t, model.Metadata.Tags.IsNull(), "absent tags must be null, not an empty list")
	require.True(t, model.Metadata.RoutingTags.IsNull())
	require.Nil(t, model.Query.Breach)
	require.Nil(t, model.StateTransition)
	require.Nil(t, model.Grouping)
	require.Nil(t, model.Artifacts)

	model.Metadata.Tags = types.ListNull(types.StringType)
	model.Metadata.RoutingTags = types.ListNull(types.StringType)
	body, diags := model.toAPIPutBody(context.Background())
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"kind": "signal",
		"enabled": true,
		"time_field": "@timestamp",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"}
	}`, string(raw))
}

func TestModel_populateFromAPI_warnsWhenResponseCarriesBuilder(t *testing.T) {
	t.Parallel()

	withBuilder := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high", "builder": {"type": "threshold"}},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), withBuilder)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Len(t, diags, 1)
	require.Equal(t, diag.SeverityWarning, diags[0].Severity())
	require.Contains(t, diags[0].Detail(), "threshold")
	require.Contains(t, diags[0].Detail(), "ES|QL")
	require.Equal(t, "CPU high", model.Metadata.Name.ValueString())

	withoutBuilder := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`)
	require.Empty(t, model.populateFromAPI(context.Background(), withoutBuilder))
}

func TestModel_toAPIPutBody_neverCarriesBuilder(t *testing.T) {
	t.Parallel()

	imported := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high", "builder": {"type": "threshold"}},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`)
	var model Model
	model.populateFromAPI(context.Background(), imported)

	body, diags := model.toAPIPutBody(context.Background())
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	require.Nil(t, body.Metadata.Builder)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "builder")
}

func TestModel_populateFromAPI_unknownNoDataStrategyIsError(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "shrug"}
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), resp)

	require.True(t, diags.HasError())
	require.Contains(t, diags[0].Detail(), "shrug")
}

func TestModel_populateFromAPI_toleratesStoredAlertNoDataStrategy(t *testing.T) {
	t.Parallel()

	resp := apiRuleResponse(t, `{
		"id": "my-rule",
		"kind": "alert",
		"enabled": true,
		"version": 1,
		"created_at": "2026-01-02T03:04:05.000Z",
		"updated_at": "2026-01-02T03:04:05.000Z",
		"metadata": {"name": "CPU high"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-*"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "alert", "query": "FROM logs-* | STATS c = COUNT()"}
	}`)

	var model Model
	diags := model.populateFromAPI(context.Background(), resp)

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.Equal(t, "alert", model.NoData.Strategy.ValueString())
	require.Equal(t, "FROM logs-* | STATS c = COUNT()", model.NoData.Query.ValueString())
}
