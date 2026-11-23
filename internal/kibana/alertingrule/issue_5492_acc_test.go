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

package alertingrule_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/google/uuid"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccReproduceIssue5492 reproduces https://github.com/elastic/terraform-provider-elasticstack/issues/5492:
// enabling a never-run elasticstack_kibana_alerting_rule (created with enabled = false, so Kibana never
// scheduled a Task Manager task for it) fails with "Provider produced inconsistent result after apply" on
// scheduled_task_id. The schema's scheduled_task_id attribute uses an unconditional
// stringplanmodifier.UseStateForUnknown(), so the plan keeps the prior null value even though flipping
// enabled to true causes Kibana to schedule a new task and return a non-null scheduled_task_id on the
// post-apply read.
func TestAccReproduceIssue5492(t *testing.T) {
	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	ruleID := uuid.New().String()

	config := func(enabled bool) string {
		return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test" {
  name        = %q
  rule_id     = %q
  consumer    = "alerts"
  notify_when = "onActiveAlert"
  params = jsonencode({
    index               = ["logs-*"]
    timeField           = "@timestamp"
    aggType             = "count"
    groupBy             = "all"
    timeWindowSize      = 30
    timeWindowUnit      = "m"
    thresholdComparator = "<"
    threshold           = [1]
  })
  rule_type_id = ".index-threshold"
  interval     = "5m"
  enabled      = %t
}
`, ruleName, ruleID, enabled)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				// Create the rule disabled; Kibana never schedules a task, so
				// scheduled_task_id is null in state.
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   config(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "enabled", "false"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_alerting_rule.test", "scheduled_task_id"),
				),
			},
			{
				// Enabling the rule causes Kibana to schedule a task and return a
				// non-null scheduled_task_id, but the plan keeps the prior null
				// value because scheduled_task_id's UseStateForUnknown plan
				// modifier is not conditioned on the enabled transition.
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   config(true),
				ExpectError:              regexp.MustCompile(`(?s)Provider produced inconsistent result after apply.*scheduled_task_id`),
			},
		},
	})
}
