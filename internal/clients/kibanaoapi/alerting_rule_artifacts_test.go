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

func Test_buildUpdateRequestBody_omitsArtifactsWhenNil(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "n",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{"index": []any{"i"}},
	}

	body, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"artifacts"`)
}

func Test_buildUpdateRequestBody_includesArtifactsWhenSet(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "n",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{"index": []any{"i"}},
		Artifacts: &models.AlertingRuleArtifacts{
			Dashboards: []models.AlertingRuleArtifactDashboard{{ID: "dash-1"}},
			InvestigationGuide: &models.AlertingRuleArtifactInvestigationGuide{
				Blob: "runbook",
			},
		},
	}

	body, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"artifacts"`)
	require.Contains(t, string(raw), `"dash-1"`)
	require.Contains(t, string(raw), `"runbook"`)
}

func Test_ConvertResponseToModel_mapsArtifacts(t *testing.T) {
	resp := map[string]any{
		"id":           "rule-1",
		"name":         "n",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"enabled":      true,
		"schedule":     map[string]any{"interval": "1m"},
		"params":       map[string]any{},
		"execution_status": map[string]any{
			"status": "",
		},
		"artifacts": map[string]any{
			"dashboards": []any{map[string]any{"id": "d1"}},
			"investigation_guide": map[string]any{
				"blob": "guide text",
			},
		},
	}

	model, diags := ConvertResponseToModel("default", resp)
	require.False(t, diags.HasError())
	require.NotNil(t, model.Artifacts)
	require.Len(t, model.Artifacts.Dashboards, 1)
	require.Equal(t, "d1", model.Artifacts.Dashboards[0].ID)
	require.NotNil(t, model.Artifacts.InvestigationGuide)
	require.Equal(t, "guide text", model.Artifacts.InvestigationGuide.Blob)
}
