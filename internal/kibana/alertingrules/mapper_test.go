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

package alertingrules

import (
	"context"
	"testing"
	"time"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestRuleElement_executionFieldsUseUTCWithThreeFractionalDigits(t *testing.T) {
	t.Parallel()

	when, err := time.Parse(time.RFC3339, "2026-09-18T16:15:41.450Z")
	require.NoError(t, err)
	status := "ok"
	element := ruleElementFromAPI(models.AlertingRule{
		ExecutionStatus: models.AlertingRuleExecutionStatus{
			Status:            &status,
			LastExecutionDate: &when,
		},
	})

	require.Equal(t, "ok", element.LastExecutionStatus.ValueString())
	require.Equal(t, "2026-09-18T16:15:41.450Z", element.LastExecutionDate.ValueString())
}

func TestRuleElement_nonUTCExecutionDateIsNormalised(t *testing.T) {
	t.Parallel()

	when, err := time.Parse(time.RFC3339, "2026-09-18T18:15:41.450+02:00")
	require.NoError(t, err)
	element := ruleElementFromAPI(models.AlertingRule{
		ExecutionStatus: models.AlertingRuleExecutionStatus{LastExecutionDate: &when},
	})

	require.Equal(t, "2026-09-18T16:15:41.450Z", element.LastExecutionDate.ValueString())
}

func TestRuleElement_neverExecutedDateIsNull(t *testing.T) {
	t.Parallel()

	element := ruleElementFromAPI(models.AlertingRule{})
	require.True(t, element.LastExecutionDate.IsNull())
}

func TestRuleElement_absentExecutionStatusIsNull(t *testing.T) {
	t.Parallel()

	element := ruleElementFromAPI(models.AlertingRule{})
	require.True(t, element.LastExecutionStatus.IsNull())
}

func TestRuleElement_unparseableExecutionDateIsNull(t *testing.T) {
	t.Parallel()

	rule, diags := kibanaoapi.ConvertResponseToModel("default", map[string]any{
		"id":           "abc",
		"name":         "rule",
		"consumer":     "alerts",
		"rule_type_id": ".index-threshold",
		"enabled":      true,
		"schedule":     map[string]any{"interval": "1m"},
		"execution_status": map[string]any{
			"last_execution_date": "not-a-timestamp",
			"status":              "ok",
		},
	})
	require.False(t, diags.HasError())
	require.NotNil(t, rule)

	element := ruleElementFromAPI(*rule)
	require.True(t, element.LastExecutionDate.IsNull())
	require.Equal(t, "ok", element.LastExecutionStatus.ValueString())
}

func TestRuleElement_absentScheduledTaskIDIsNull(t *testing.T) {
	t.Parallel()

	element := ruleElementFromAPI(models.AlertingRule{})
	require.True(t, element.ScheduledTaskID.IsNull())
}

func TestRuleElement_noTagsIsEmptySet(t *testing.T) {
	t.Parallel()

	element := ruleElementFromAPI(models.AlertingRule{})
	require.False(t, element.Tags.IsNull())
	require.Empty(t, element.Tags.Elements())
}

func TestRuleElement_identityAndConfigurationFields(t *testing.T) {
	t.Parallel()

	enabled := true
	taskID := "task-1"
	element := ruleElementFromAPI(models.AlertingRule{
		RuleID:          "abc",
		Name:            "a-rule",
		RuleTypeID:      ".index-threshold",
		Consumer:        "alerts",
		Enabled:         &enabled,
		Tags:            []string{"first", "second"},
		ScheduledTaskID: &taskID,
	})

	require.Equal(t, "abc", element.ID.ValueString())
	require.Equal(t, "a-rule", element.Name.ValueString())
	require.Equal(t, ".index-threshold", element.RuleTypeID.ValueString())
	require.Equal(t, "alerts", element.Consumer.ValueString())
	require.True(t, element.Enabled.ValueBool())
	require.Equal(t, "task-1", element.ScheduledTaskID.ValueString())
	require.ElementsMatch(t, []string{"first", "second"}, setStrings(t, element.Tags))
}

func setStrings(t *testing.T, set types.Set) []string {
	t.Helper()
	var values []string
	require.False(t, set.ElementsAs(context.Background(), &values, false).HasError())
	return values
}
