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

package kibanaoapi

import (
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateRequestBody_IncludesArtifacts(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		Artifacts: &models.AlertingRuleArtifacts{
			Dashboards: []models.AlertingRuleArtifactDashboard{{ID: "dashboard-1"}},
			InvestigationGuide: &models.AlertingRuleArtifactInvestigationGuide{
				Blob: "guide",
			},
		},
	}

	body := buildCreateRequestBody(rule)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"artifacts"`)
	require.Contains(t, string(raw), `"dashboards":[{"id":"dashboard-1"}]`)
	require.Contains(t, string(raw), `"investigation_guide":{"blob":"guide"}`)
}

func TestBuildUpdateRequestBody_IncludesEmptyArtifactsDashboardsList(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		Artifacts: &models.AlertingRuleArtifacts{
			Dashboards: []models.AlertingRuleArtifactDashboard{},
		},
	}

	body := buildUpdateRequestBody(rule)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"artifacts"`)
	require.Contains(t, string(raw), `"dashboards":[]`)
}

func TestConvertResponseToModel_PreservesEmptyArtifactsDashboards(t *testing.T) {
	resp := map[string]any{
		"id":           "id",
		"name":         "name",
		"consumer":     "consumer",
		"params":       map[string]any{},
		"rule_type_id": "rule-type-id",
		"enabled":      true,
		"schedule": map[string]any{
			"interval": "1m",
		},
		"artifacts": map[string]any{
			"dashboards": []any{},
		},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.NotNil(t, model)
	require.NotNil(t, model.Artifacts)
	require.NotNil(t, model.Artifacts.Dashboards)
	require.Empty(t, model.Artifacts.Dashboards)
}

func TestConvertResponseToModel_PopulatesArtifactsDashboardsAndGuide(t *testing.T) {
	resp := map[string]any{
		"id":           "id",
		"name":         "name",
		"consumer":     "consumer",
		"params":       map[string]any{},
		"rule_type_id": "rule-type-id",
		"enabled":      true,
		"schedule": map[string]any{
			"interval": "1m",
		},
		"artifacts": map[string]any{
			"dashboards": []any{
				map[string]any{"id": "dashboard-1"},
				map[string]any{"id": "dashboard-2"},
			},
			"investigation_guide": map[string]any{
				"blob": "guide",
			},
		},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	require.NotNil(t, model)
	require.NotNil(t, model.Artifacts)
	require.Len(t, model.Artifacts.Dashboards, 2)
	require.Equal(t, "dashboard-1", model.Artifacts.Dashboards[0].ID)
	require.Equal(t, "dashboard-2", model.Artifacts.Dashboards[1].ID)
	require.NotNil(t, model.Artifacts.InvestigationGuide)
	require.Equal(t, "guide", model.Artifacts.InvestigationGuide.Blob)
}
