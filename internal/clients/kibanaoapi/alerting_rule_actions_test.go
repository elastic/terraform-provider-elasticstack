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

func Test_buildActionsSlice_empty(t *testing.T) {
	result := buildActionsSlice(nil)
	require.Empty(t, result)

	result = buildActionsSlice([]models.AlertingRuleAction{})
	require.Empty(t, result)
}

func Test_buildActionsSlice_basicFields(t *testing.T) {
	actions := []models.AlertingRuleAction{
		{
			ID:     "connector-1",
			Group:  "default",
			Params: map[string]any{"message": "alert fired"},
		},
	}
	result := buildActionsSlice(actions)
	require.Len(t, result, 1)
	require.Equal(t, "connector-1", result[0].Id)
	require.NotNil(t, result[0].Group)
	require.Equal(t, "default", *result[0].Group)
	require.NotNil(t, result[0].Params)
}

func Test_buildActionsSlice_omitsGroupWhenEmpty(t *testing.T) {
	actions := []models.AlertingRuleAction{
		{ID: "connector-1", Group: ""},
	}
	result := buildActionsSlice(actions)
	require.Len(t, result, 1)
	require.Nil(t, result[0].Group)
}

func Test_buildActionsSlice_frequency(t *testing.T) {
	throttle := "10s"
	actions := []models.AlertingRuleAction{
		{
			ID:    "connector-1",
			Group: "default",
			Frequency: &models.ActionFrequency{
				NotifyWhen: "onThrottleInterval",
				Summary:    true,
				Throttle:   &throttle,
			},
		},
	}
	result := buildActionsSlice(actions)
	require.Len(t, result, 1)
	require.NotNil(t, result[0].Frequency)
	require.Equal(t, "onThrottleInterval", result[0].Frequency.NotifyWhen)
	require.True(t, result[0].Frequency.Summary)
	require.NotNil(t, result[0].Frequency.Throttle)
	require.Equal(t, throttle, *result[0].Frequency.Throttle)
}

func Test_buildActionsSlice_alertsFilter(t *testing.T) {
	kql := "tags: (foo)"
	actions := []models.AlertingRuleAction{
		{
			ID:    "connector-1",
			Group: "default",
			AlertsFilter: &models.ActionAlertsFilter{
				Kql: &kql,
				Timeframe: &models.AlertsFilterTimeframe{
					Days:       []int32{1, 3, 5},
					Timezone:   "Europe/Berlin",
					HoursStart: "08:00",
					HoursEnd:   "17:00",
				},
			},
		},
	}
	result := buildActionsSlice(actions)
	require.Len(t, result, 1)
	require.NotNil(t, result[0].AlertsFilter)
	require.NotNil(t, result[0].AlertsFilter.Query)
	require.Equal(t, kql, result[0].AlertsFilter.Query.Kql)
	require.Empty(t, result[0].AlertsFilter.Query.Filters)

	require.NotNil(t, result[0].AlertsFilter.Timeframe)
	require.Equal(t, []int{1, 3, 5}, result[0].AlertsFilter.Timeframe.Days)
	require.Equal(t, "Europe/Berlin", result[0].AlertsFilter.Timeframe.Timezone)
	require.Equal(t, "08:00", result[0].AlertsFilter.Timeframe.Hours.Start)
	require.Equal(t, "17:00", result[0].AlertsFilter.Timeframe.Hours.End)
}

// Test_buildCreateAndUpdateActionsJSON verifies that buildCreateRequestBody and
// buildUpdateRequestBody produce identical JSON for the actions array, confirming
// that the shared buildActionsSlice helper produces consistent output across both paths.
func Test_buildCreateAndUpdateActionsJSON(t *testing.T) {
	throttle := "30s"
	kql := "host.name: webserver"
	rule := models.AlertingRule{
		Name:       "test-rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "5m"},
		Params:     map[string]any{},
		Actions: []models.AlertingRuleAction{
			{
				ID:    "connector-id-1",
				Group: "threshold met",
				Frequency: &models.ActionFrequency{
					NotifyWhen: "onThrottleInterval",
					Summary:    false,
					Throttle:   &throttle,
				},
				AlertsFilter: &models.ActionAlertsFilter{
					Kql: &kql,
					Timeframe: &models.AlertsFilterTimeframe{
						Days:       []int32{2, 4},
						Timezone:   "UTC",
						HoursStart: "09:00",
						HoursEnd:   "18:00",
					},
				},
			},
			{
				ID:    "connector-id-2",
				Group: "recovered",
			},
		},
	}

	createBody, err := buildCreateRequestBody(rule)
	require.NoError(t, err)

	updateBody, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)

	createActionsJSON, err := json.Marshal(createBody.Actions)
	require.NoError(t, err)

	updateActionsJSON, err := json.Marshal(updateBody.Actions)
	require.NoError(t, err)

	// Unmarshal both into a generic structure for field-level comparison
	var createActions, updateActions []map[string]any
	require.NoError(t, json.Unmarshal(createActionsJSON, &createActions))
	require.NoError(t, json.Unmarshal(updateActionsJSON, &updateActions))

	require.Len(t, updateActions, len(createActions))
	for i := range createActions {
		require.Equal(t, createActions[i], updateActions[i], "action[%d] differs between create and update", i)
	}
}

func Test_buildRequestBody_returnsErrorOnInvalidActionParams(t *testing.T) {
	tests := []struct {
		name  string
		build func(models.AlertingRule) error
	}{
		{
			name: "create",
			build: func(rule models.AlertingRule) error {
				_, err := buildCreateRequestBody(rule)
				return err
			},
		},
		{
			name: "update",
			build: func(rule models.AlertingRule) error {
				_, err := buildUpdateRequestBody(rule)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := models.AlertingRule{
				Name:       "test-rule",
				Consumer:   "alerts",
				RuleTypeID: ".index-threshold",
				Schedule:   models.AlertingRuleSchedule{Interval: "5m"},
				Params:     map[string]any{},
				Actions: []models.AlertingRuleAction{
					{
						ID:     "connector-id-1",
						Group:  "threshold met",
						Params: map[string]any{"invalid": make(chan int)},
					},
				},
			}

			err := tt.build(rule)
			require.Error(t, err)
			require.ErrorContains(t, err, "convert actions")
			require.ErrorContains(t, err, "marshal actions")
			require.ErrorContains(t, err, "unsupported type: chan int")
		})
	}
}
