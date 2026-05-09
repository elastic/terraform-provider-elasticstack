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

func Test_buildOptionalRuleFields_omitsAllWhenNil(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
	}
	fields := buildOptionalRuleFields(rule)
	data, err := json.Marshal(fields)
	require.NoError(t, err)
	require.Equal(t, `{}`, string(data))
}

func Test_buildOptionalRuleFields_omitsNotifyWhenWhenEmpty(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		NotifyWhen: new(""),
	}
	fields := buildOptionalRuleFields(rule)
	require.Nil(t, fields.NotifyWhen)
}

func Test_buildOptionalRuleFields_includesAllWhenSet(t *testing.T) {
	enabled := true
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		NotifyWhen: new("onActiveAlert"),
		Throttle:   new("10s"),
		Tags:       []string{"tag1", "tag2"},
		AlertDelay: new(float32(3)),
		Flapping: &models.AlertingRuleFlapping{
			LookBackWindow:        10,
			StatusChangeThreshold: 2,
			Enabled:               &enabled,
		},
	}
	fields := buildOptionalRuleFields(rule)

	require.Equal(t, new("onActiveAlert"), fields.NotifyWhen)
	require.Equal(t, new("10s"), fields.Throttle)
	require.NotNil(t, fields.Tags)
	require.Equal(t, []string{"tag1", "tag2"}, *fields.Tags)
	require.NotNil(t, fields.AlertDelay)
	require.InEpsilon(t, float32(3), fields.AlertDelay.Active, 1e-6)
	require.NotNil(t, fields.Flapping)
	require.InEpsilon(t, float32(10), fields.Flapping.LookBackWindow, 1e-6)
	require.InEpsilon(t, float32(2), fields.Flapping.StatusChangeThreshold, 1e-6)
	require.Equal(t, &enabled, fields.Flapping.Enabled)
}

func Test_buildCreateRequestBody_omitsOptionalFieldsWhenNil(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
	}
	body, err := buildCreateRequestBody(rule)
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	s := string(raw)
	require.NotContains(t, s, `"notify_when"`)
	require.NotContains(t, s, `"throttle"`)
	require.NotContains(t, s, `"tags"`)
	require.NotContains(t, s, `"alert_delay"`)
	require.NotContains(t, s, `"flapping"`)
}

func Test_buildCreateRequestBody_includesOptionalFieldsWhenSet(t *testing.T) {
	enabled := true
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		NotifyWhen: new("onActiveAlert"),
		Throttle:   new("10s"),
		Tags:       []string{"a", "b"},
		AlertDelay: new(float32(5)),
		Flapping: &models.AlertingRuleFlapping{
			LookBackWindow:        10,
			StatusChangeThreshold: 3,
			Enabled:               &enabled,
		},
	}
	body, err := buildCreateRequestBody(rule)
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	s := string(raw)
	require.Contains(t, s, `"notify_when":"onActiveAlert"`)
	require.Contains(t, s, `"throttle":"10s"`)
	require.Contains(t, s, `"tags":["a","b"]`)
	require.Contains(t, s, `"alert_delay"`)
	require.Contains(t, s, `"active":5`)
	require.Contains(t, s, `"flapping"`)
	require.Contains(t, s, `"look_back_window":10`)
	require.Contains(t, s, `"status_change_threshold":3`)
}

func Test_buildUpdateRequestBody_omitsOptionalFieldsWhenNil(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
	}
	body, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	s := string(raw)
	require.NotContains(t, s, `"notify_when"`)
	require.NotContains(t, s, `"throttle"`)
	require.NotContains(t, s, `"tags"`)
	require.NotContains(t, s, `"alert_delay"`)
	require.NotContains(t, s, `"flapping"`)
}

func Test_buildUpdateRequestBody_includesOptionalFieldsWhenSet(t *testing.T) {
	enabled := true
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		NotifyWhen: new("onActiveAlert"),
		Throttle:   new("10s"),
		Tags:       []string{"a", "b"},
		AlertDelay: new(float32(5)),
		Flapping: &models.AlertingRuleFlapping{
			LookBackWindow:        10,
			StatusChangeThreshold: 3,
			Enabled:               &enabled,
		},
	}
	body, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	s := string(raw)
	require.Contains(t, s, `"notify_when":"onActiveAlert"`)
	require.Contains(t, s, `"throttle":"10s"`)
	require.Contains(t, s, `"tags":["a","b"]`)
	require.Contains(t, s, `"alert_delay"`)
	require.Contains(t, s, `"active":5`)
	require.Contains(t, s, `"flapping"`)
	require.Contains(t, s, `"look_back_window":10`)
	require.Contains(t, s, `"status_change_threshold":3`)
}

// Test_createAndUpdateProduceSameOptionalFieldJSON verifies that both body builders
// produce identical JSON for the shared optional fields, confirming the refactoring
// preserves parity.
func Test_createAndUpdateProduceSameOptionalFieldJSON(t *testing.T) {
	enabled := true
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		NotifyWhen: new("onActiveAlert"),
		Throttle:   new("10s"),
		Tags:       []string{"x"},
		AlertDelay: new(float32(2)),
		Flapping: &models.AlertingRuleFlapping{
			LookBackWindow:        5,
			StatusChangeThreshold: 1,
			Enabled:               &enabled,
		},
	}

	create, err := buildCreateRequestBody(rule)
	require.NoError(t, err)
	update, err := buildUpdateRequestBody(rule)
	require.NoError(t, err)

	var createMap, updateMap map[string]any
	createRaw, _ := json.Marshal(create)
	updateRaw, _ := json.Marshal(update)
	require.NoError(t, json.Unmarshal(createRaw, &createMap))
	require.NoError(t, json.Unmarshal(updateRaw, &updateMap))

	for _, field := range []string{"notify_when", "throttle", "tags", "alert_delay", "flapping"} {
		require.Equal(t, createMap[field], updateMap[field], "field %q differs between create and update bodies", field)
	}
}
