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

package rule_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const resourceAddr = "elasticstack_kibana_alerting_v2_rule.test"

// The Alerting v2 rules API first shipped in 9.6.0; the prerelease form lets
// 9.6.0-SNAPSHOT stacks run the suite.
var minAlertingV2RuleAPISupport = version.Must(version.NewVersion("9.6.0-SNAPSHOT"))

func TestAccResourceAlertingV2Rule(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID := uuid.New().String()
	ruleName := "tf-acc-" + uuid.New().String()[:8]
	configVariables := config.Variables{
		"name":    config.StringVariable(ruleName),
		"rule_id": config.StringVariable(ruleID),
	}
	var createdAt string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "id", "default/"+ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "rule_id", ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "space_id", "default"),
					resource.TestCheckResourceAttr(resourceAddr, "kind", "alert"),
					resource.TestCheckResourceAttr(resourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "time_field", "@timestamp"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.every", "1m"),
					resource.TestCheckNoResourceAttr(resourceAddr, "schedule.lookback"),
					resource.TestCheckResourceAttr(resourceAddr, "query.base", "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.strategy", "no_breach"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.strategy", "ignore"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
					resource.TestCheckResourceAttrSet(resourceAddr, "updated_at"),
					rememberAttr(resourceAddr, "created_at", &createdAt),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName+"-updated"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.every", "2m"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "2"),
					resource.TestCheckResourceAttrPtr(resourceAddr, "created_at", &createdAt),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          configVariables,
				ResourceName:             resourceAddr,
				ImportState:              true,
				ImportStateId:            "default/" + ruleID,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_kind"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceAddr, plancheck.ResourceActionDestroyBeforeCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "kind", "signal"),
					resource.TestCheckNoResourceAttr(resourceAddr, "recovery"),
					resource.TestCheckNoResourceAttr(resourceAddr, "no_data"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
				),
			},
		},
	})
}

func emptyPlanAfterApply() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
}

func rememberAttr(addr, key string, into *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %s not found in state", addr)
		}
		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute %s not set on %s", key, addr)
		}
		*into = value
		return nil
	}
}

var checkAlertingV2RuleDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_alerting_v2_rule",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, ruleID string) (bool, error) {
		rule, diags := kibanaoapi.GetAlertingV2Rule(ctx, client, spaceID, ruleID)
		if diags.HasError() {
			return false, diagutil.FwDiagsAsError(diags)
		}
		return rule != nil, nil
	},
)
