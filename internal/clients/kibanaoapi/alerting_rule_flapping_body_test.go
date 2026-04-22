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

func Test_buildUpdateRequestBody_omitsFlappingWhenNil(t *testing.T) {
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		Flapping:   nil,
	}
	body := buildUpdateRequestBody(rule)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.NotContains(t, string(raw), `"flapping"`)
}

func Test_buildUpdateRequestBody_includesFlappingWhenSet(t *testing.T) {
	enabled := true
	rule := models.AlertingRule{
		Name:       "rule",
		Consumer:   "alerts",
		RuleTypeID: ".index-threshold",
		Schedule:   models.AlertingRuleSchedule{Interval: "1m"},
		Params:     map[string]any{},
		Flapping: &models.AlertingRuleFlapping{
			LookBackWindow:        10,
			StatusChangeThreshold: 2,
			Enabled:               &enabled,
		},
	}
	body := buildUpdateRequestBody(rule)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"flapping"`)
	require.Contains(t, string(raw), `"look_back_window"`)
	require.Contains(t, string(raw), `"status_change_threshold"`)
}
