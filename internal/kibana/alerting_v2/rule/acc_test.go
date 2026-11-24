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
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
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

	ruleID, ruleName, configVariables := ruleTestVariables()
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
				ConfigPlanChecks:         expectActionThenEmptyPlan(plancheck.ResourceActionDestroyBeforeCreate),
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

func TestAccResourceAlertingV2Rule_full(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()

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
					resource.TestCheckResourceAttr(resourceAddr, "kind", "alert"),
					resource.TestCheckResourceAttr(resourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "time_field", "event.ingested"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.description", "Hosts with an unusual volume of log events"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.0", "z"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.1", "a"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.0", "pager"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.1", "ops"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.every", "1m"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.lookback", "5m"),
					resource.TestCheckResourceAttr(resourceAddr, "query.base", "FROM logs-* | STATS count = COUNT(*) BY host.name"),
					resource.TestCheckResourceAttr(resourceAddr, "query.breach.segment", "WHERE count > 500"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.strategy", "condition"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.segment", "WHERE count < 50"),
					resource.TestCheckNoResourceAttr(resourceAddr, "recovery.query"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.strategy", "keep_last"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.query", "FROM logs-* | STATS c = COUNT() BY host.name"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.count", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.timeframe", "5m"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.operator", "and"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.recovering.count", "1"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition.recovering.timeframe"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition.recovering.operator"),
					resource.TestCheckResourceAttr(resourceAddr, "grouping.fields.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "grouping.fields.0", "host.name"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.id", "rb"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.type", "runbook"),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.0.data", `{"content": "# Steps"}`),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.1.id", "db"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.1.type", "dashboard"),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.1.data", `{"dashboard_id": "ops-overview"}`),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("reformat"),
				ConfigVariables:          configVariables,
				PlanOnly:                 true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName+"-updated"),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.0.data", `{"content": "# Steps - check the host first"}`),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.1.data", `{"dashboard_id": "ops-overview"}`),
					resource.TestCheckResourceAttr(resourceAddr, "version", "2"),
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
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_optionals"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceAddr, "grouping"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition"),
					resource.TestCheckNoResourceAttr(resourceAddr, "artifacts"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "3"),
				),
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_signal(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()

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
					resource.TestCheckResourceAttr(resourceAddr, "kind", "signal"),
					resource.TestCheckResourceAttr(resourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "time_field", "@timestamp"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
					resource.TestCheckNoResourceAttr(resourceAddr, "metadata.description"),
					resource.TestCheckNoResourceAttr(resourceAddr, "metadata.tags"),
					resource.TestCheckNoResourceAttr(resourceAddr, "metadata.routing_tags"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.every", "1m"),
					resource.TestCheckResourceAttr(resourceAddr, "query.base", "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"),
					resource.TestCheckNoResourceAttr(resourceAddr, "query.breach"),
					resource.TestCheckNoResourceAttr(resourceAddr, "recovery"),
					resource.TestCheckNoResourceAttr(resourceAddr, "no_data"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition"),
					resource.TestCheckNoResourceAttr(resourceAddr, "grouping"),
					resource.TestCheckNoResourceAttr(resourceAddr, "artifacts"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ResourceName:             resourceAddr,
				ImportState:              true,
				ImportStateId:            "default/" + ruleID,
				ImportStateVerify:        true,
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_enabledToggle(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	_, _, configVariables := ruleTestVariables()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("disabled"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("enabled"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "2"),
				),
			},
		},
	})
}

// Server-side rejections the provider does not pre-empt: ES|QL syntax, the
// deployment's minimum schedule interval (1m by default, above the schema's
// 5s floor), and type-specific artifact data. Each surfaces the server's code
// and message, and nothing is created.
func TestAccResourceAlertingV2Rule_serverErrors(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	_, _, configVariables := ruleTestVariables()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_esql"),
				ConfigVariables:          configVariables,
				ExpectError:              regexp.MustCompile(`(?s)HTTP 400 BAD_REQUEST.*query\.base: Invalid ES\|QL query`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("below_deployment_minimum"),
				ConfigVariables:          configVariables,
				ExpectError:              regexp.MustCompile(`(?s)HTTP 400 SCHEDULE_INTERVAL_TOO_SHORT.*"10s" is shorter than the allowed\s+minimum`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("invalid_artifact_data"),
				ConfigVariables:          configVariables,
				ExpectError:              regexp.MustCompile(`(?s)HTTP 400 INVALID_ARTIFACT_DATA.*Artifact "rb" of type "runbook" has invalid data`),
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_space(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()
	spaceID := "tf-acc-" + uuid.New().String()[:8]
	newSpaceID := "tf-acc-" + uuid.New().String()[:8]
	configVariables["space_id"] = config.StringVariable(spaceID)

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
					resource.TestCheckResourceAttr(resourceAddr, "id", spaceID+"/"+ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceAddr, "rule_id", ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
					checkRuleInSpace(spaceID, &ruleID),
					checkRuleAbsentFromSpace("default", &ruleID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ResourceName:             resourceAddr,
				ImportState:              true,
				ImportStateId:            spaceID + "/" + ruleID,
				ImportStateVerify:        true,
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("change_space"),
				ConfigVariables: config.Variables{
					"name":         config.StringVariable(ruleName),
					"rule_id":      config.StringVariable(ruleID),
					"space_id":     config.StringVariable(spaceID),
					"new_space_id": config.StringVariable(newSpaceID),
				},
				ConfigPlanChecks: expectActionThenEmptyPlan(plancheck.ResourceActionDestroyBeforeCreate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "id", newSpaceID+"/"+ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "space_id", newSpaceID),
					checkRuleInSpace(newSpaceID, &ruleID),
					checkRuleAbsentFromSpace(spaceID, &ruleID),
				),
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_ruleID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	_, ruleName, _ := ruleTestVariables()
	suppliedRuleID := "cpu-high_" + uuid.New().String()[:8]
	nameOnly := config.Variables{"name": config.StringVariable(ruleName)}
	var generatedRuleID string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("generated"),
				ConfigVariables:          nameOnly,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceAddr, "rule_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					rememberAttr(resourceAddr, "rule_id", &generatedRuleID),
					checkCompositeID(resourceAddr, "default", &generatedRuleID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("generated"),
				ConfigVariables:          nameOnly,
				ConfigPlanChecks:         expectNoOpApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceAddr, "rule_id", &generatedRuleID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("supplied"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(ruleName),
					"rule_id": config.StringVariable(suppliedRuleID),
				},
				ConfigPlanChecks: expectActionThenEmptyPlan(plancheck.ResourceActionDestroyBeforeCreate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "rule_id", suppliedRuleID),
					resource.TestCheckResourceAttr(resourceAddr, "id", "default/"+suppliedRuleID),
					checkRuleInSpace("default", &suppliedRuleID),
					checkRuleAbsentFromSpace("default", &generatedRuleID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("supplied"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(ruleName),
					"rule_id": config.StringVariable(suppliedRuleID),
				},
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateId:     "default/" + suppliedRuleID,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_externalDelete(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, _, configVariables := ruleTestVariables()
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
				Check:                    rememberAttr(resourceAddr, "created_at", &createdAt),
			},
			{
				PreConfig:                func() { deleteRule(t, ruleID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigVariables:          configVariables,
				RefreshState:             true,
				ExpectNonEmptyPlan:       true,
				Check:                    checkResourceAbsent(resourceAddr),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         expectActionThenEmptyPlan(plancheck.ResourceActionCreate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "id", "default/"+ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
					checkAttrDiffers(resourceAddr, "created_at", &createdAt),
				),
			},
			{
				PreConfig:                func() { deleteRule(t, ruleID) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				Destroy:                  true,
			},
		},
	})
}

func TestAccResourceAlertingV2Rule_existingID(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()
	unmanagedName := ruleName + "-unmanaged"
	unmanagedBody := `{
		"kind": "alert",
		"metadata": {"name": "` + unmanagedName + `"},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			t.Cleanup(func() { deleteRule(t, ruleID) })
		},
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { putRuleJSON(t, "default", ruleID, unmanagedBody) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ExpectError:              regexp.MustCompile(`(?s)rule\s+already\s+exists.*"default/` + ruleID + `".*terraform\s+import`),
			},
			{
				PreConfig: func() {
					unmanaged := getRule(t, "default", ruleID)
					if unmanaged == nil {
						t.Fatalf("unmanaged rule %s disappeared during the failed create", ruleID)
					}
					if unmanaged.Version != 1 || unmanaged.Metadata.Name != unmanagedName {
						t.Fatalf("unmanaged rule %s was written by the failed create: %+v", ruleID, unmanaged)
					}
					deleteRule(t, ruleID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         expectActionThenEmptyPlan(plancheck.ResourceActionCreate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "id", "default/"+ruleID),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
				),
			},
		},
	})
}

// The Kibana UI stores metadata.builder on rules it authors. The resource does
// not manage it: a matching configuration leaves it in place and the first
// write that changes the rule clears it. The warning Read emits while the
// builder is present is asserted by the Read-path unit tests; the acceptance
// framework does not expose warning diagnostics.
func TestAccResourceAlertingV2Rule_builderImport(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()
	builderBody := `{
		"kind": "alert",
		"metadata": {"name": "` + ruleName + `", "builder": {"type": "threshold"}},
		"schedule": {"every": "1m"},
		"query": {"base": "FROM logs-* | STATS count = COUNT(*) BY host.name | WHERE count > 100"},
		"recovery": {"strategy": "no_breach"},
		"no_data": {"strategy": "ignore"}
	}`

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			t.Cleanup(func() { deleteRule(t, ruleID) })
		},
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig:                func() { putRuleJSON(t, "default", ruleID, builderBody) },
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("matching"),
				ConfigVariables:          configVariables,
				ResourceName:             resourceAddr,
				ImportState:              true,
				ImportStateId:            "default/" + ruleID,
				ImportStatePersist:       true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported instance, got %d", len(states))
					}
					for key, want := range map[string]string{
						"id":            "default/" + ruleID,
						"space_id":      "default",
						"rule_id":       ruleID,
						"metadata.name": ruleName,
						"version":       "1",
					} {
						if got := states[0].Attributes[key]; got != want {
							return fmt.Errorf("imported %s = %q, want %q", key, got, want)
						}
					}
					return nil
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("matching"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         expectNoOpApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
					checkServerBuilderType("default", ruleID, "threshold"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("changed"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         expectActionThenEmptyPlan(plancheck.ResourceActionUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName+"-updated"),
					resource.TestCheckResourceAttr(resourceAddr, "version", "2"),
					checkServerBuilderAbsent("default", ruleID),
				),
			},
		},
	})
}

// A resource-level kibana_connection must be the client every lifecycle
// operation goes through: the rule is created through the real endpoint and
// an unreachable endpoint in the block fails the apply, which could not happen
// if the provider-level client were used instead.
func TestAccResourceAlertingV2Rule_kibanaConnection(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, ruleName, configVariables := ruleTestVariables()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheckWithExplicitKibanaEndpoint(t)
			acctest.PreCheck(t)
		},
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          acctest.KibanaConnectionVariables(configVariables),
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					append(
						acctest.KibanaConnectionAuthChecks(resourceAddr),
						resource.TestCheckResourceAttr(resourceAddr, "kibana_connection.#", "1"),
						resource.TestCheckResourceAttr(resourceAddr, "kibana_connection.0.endpoints.0", acctest.KibanaConnectionEndpoint()),
						resource.TestCheckResourceAttr(resourceAddr, "id", "default/"+ruleID),
						resource.TestCheckResourceAttr(resourceAddr, "metadata.name", ruleName),
						checkRuleInSpace("default", &ruleID),
					)...,
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: acctest.KibanaConnectionVariables(configVariables, config.Variables{
					"kibana_endpoints": config.ListVariable(config.StringVariable("http://127.0.0.1:1")),
				}),
				ExpectError: regexp.MustCompile(`(?is)connection refused|dial|unable|failed`),
			},
		},
	})
}

// Updates every optional field, covering the recovery and no-data strategies
// and the recovering phase the other suites leave out.
func TestAccResourceAlertingV2Rule_optionalFieldUpdates(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minAlertingV2RuleAPISupport, versionutils.FlavorAny)

	ruleID, _, configVariables := ruleTestVariables()
	var updatedAt string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingV2RuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("recovery_query"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "time_field", "event.ingested"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.description", "Hosts whose error rate stays high"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.0", "errors"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.0", "ops"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.lookback", "5m"),
					resource.TestCheckResourceAttr(resourceAddr, "query.base", "FROM logs-* | STATS errors = COUNT(*) BY host.name"),
					resource.TestCheckResourceAttr(resourceAddr, "query.breach.segment", "WHERE errors > 100"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.strategy", "query"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.query", "FROM logs-* | STATS errors = COUNT(*) BY host.name | WHERE errors < 10"),
					resource.TestCheckNoResourceAttr(resourceAddr, "recovery.segment"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.strategy", "resolve"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.query", "FROM logs-* | STATS seen = COUNT(*) BY host.name"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.count", "2"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition.pending.timeframe"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition.pending.operator"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.recovering.count", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.recovering.timeframe", "5m"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.recovering.operator", "or"),
					resource.TestCheckResourceAttr(resourceAddr, "grouping.fields.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.id", "rb"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.type", "runbook"),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.0.data", `{"content": "# Check the host"}`),
					resource.TestCheckResourceAttr(resourceAddr, "version", "1"),
					rememberAttr(resourceAddr, "updated_at", &updatedAt),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("recovery_manual"),
				ConfigVariables:          configVariables,
				ConfigPlanChecks:         emptyPlanAfterApply(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddr, "time_field", "event.created"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.description", "Services whose error rate stays high"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.tags.1", "services"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "metadata.routing_tags.0", "pager"),
					resource.TestCheckResourceAttr(resourceAddr, "schedule.lookback", "10m"),
					resource.TestCheckResourceAttr(resourceAddr, "query.base", "FROM logs-* | STATS errors = COUNT(*) BY host.name, service.name"),
					resource.TestCheckResourceAttr(resourceAddr, "query.breach.segment", "WHERE errors > 250"),
					resource.TestCheckResourceAttr(resourceAddr, "recovery.strategy", "manual"),
					resource.TestCheckNoResourceAttr(resourceAddr, "recovery.query"),
					resource.TestCheckResourceAttr(resourceAddr, "no_data.strategy", "resolve"),
					resource.TestCheckNoResourceAttr(resourceAddr, "no_data.query"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.count", "3"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.timeframe", "10m"),
					resource.TestCheckResourceAttr(resourceAddr, "state_transition.pending.operator", "or"),
					resource.TestCheckNoResourceAttr(resourceAddr, "state_transition.recovering"),
					resource.TestCheckResourceAttr(resourceAddr, "grouping.fields.#", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "grouping.fields.1", "service.name"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.id", "db"),
					resource.TestCheckResourceAttr(resourceAddr, "artifacts.0.type", "dashboard"),
					checks.TestCheckResourceAttrJSONSubset(resourceAddr, "artifacts.0.data", `{"dashboard_id": "service-errors"}`),
					resource.TestCheckResourceAttr(resourceAddr, "version", "2"),
					checkAttrDiffers(resourceAddr, "updated_at", &updatedAt),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("recovery_manual"),
				ConfigVariables:          configVariables,
				ResourceName:             resourceAddr,
				ImportState:              true,
				ImportStateId:            "default/" + ruleID,
				ImportStateVerify:        true,
			},
		},
	})
}

// ruleTestVariables returns a fresh rule id and unique name together with the
// config variables every rule test module declares.
func ruleTestVariables() (ruleID, ruleName string, configVariables config.Variables) {
	ruleID = uuid.New().String()
	ruleName = "tf-acc-" + uuid.New().String()[:8]
	configVariables = config.Variables{
		"name":    config.StringVariable(ruleName),
		"rule_id": config.StringVariable(ruleID),
	}
	return ruleID, ruleName, configVariables
}

func emptyPlanAfterApply() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
}

func expectNoOpApply() resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
		PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}
}

func expectActionThenEmptyPlan(action plancheck.ResourceActionType) resource.ConfigPlanChecks {
	return resource.ConfigPlanChecks{
		PreApply:             []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceAddr, action)},
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

func acceptanceKibanaClient() (*kibanaoapi.Client, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	return client.GetKibanaOapiClient(), nil
}

// lookupRule reads the rule through a direct space-routed GET outside the
// provider. A nil rule means the server does not hold it in that space.
func lookupRule(spaceID, ruleID string) (*kbapi.KibanaHTTPAPIsAlertingRuleResponse, error) {
	client, err := acceptanceKibanaClient()
	if err != nil {
		return nil, err
	}
	rule, diags := kibanaoapi.GetAlertingV2Rule(context.Background(), client, spaceID, ruleID)
	if diags.HasError() {
		return nil, diagutil.FwDiagsAsError(diags)
	}
	return rule, nil
}

// checkRuleInSpace and checkRuleAbsentFromSpace prove the /s/<space> request
// routing. ruleID is dereferenced at check time so ids generated by an earlier
// step can be used.
func checkRuleInSpace(spaceID string, ruleID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		rule, err := lookupRule(spaceID, *ruleID)
		if err != nil {
			return err
		}
		if rule == nil {
			return fmt.Errorf("rule %s/%s not found", spaceID, *ruleID)
		}
		return nil
	}
}

func checkRuleAbsentFromSpace(spaceID string, ruleID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		rule, err := lookupRule(spaceID, *ruleID)
		if err != nil {
			return err
		}
		if rule != nil {
			return fmt.Errorf("rule %s/%s still exists", spaceID, *ruleID)
		}
		return nil
	}
}

func checkCompositeID(addr, spaceID string, ruleID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return resource.TestCheckResourceAttr(addr, "id", spaceID+"/"+*ruleID)(s)
	}
}

func getRule(t *testing.T, spaceID, ruleID string) *kbapi.KibanaHTTPAPIsAlertingRuleResponse {
	t.Helper()
	rule, err := lookupRule(spaceID, ruleID)
	if err != nil {
		t.Fatalf("failed to get rule %s/%s: %v", spaceID, ruleID, err)
	}
	return rule
}

// putRuleJSON writes a rule outside Terraform through the Alerting v2 API so
// tests can stage unmanaged or UI-authored rules.
func putRuleJSON(t *testing.T, spaceID, ruleID, bodyJSON string) {
	t.Helper()
	var body kbapi.KibanaHTTPAPIsAlertingPutRule
	if err := json.Unmarshal([]byte(bodyJSON), &body); err != nil {
		t.Fatalf("invalid rule body: %v", err)
	}
	client, err := acceptanceKibanaClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}
	if _, diags := kibanaoapi.PutAlertingV2Rule(context.Background(), client, spaceID, ruleID, body); diags.HasError() {
		t.Fatalf("failed to put rule %s/%s: %v", spaceID, ruleID, diags)
	}
}

func serverBuilderType(spaceID, ruleID string) (string, bool, error) {
	rule, err := lookupRule(spaceID, ruleID)
	if err != nil {
		return "", false, err
	}
	if rule == nil {
		return "", false, fmt.Errorf("rule %s/%s not found", spaceID, ruleID)
	}
	if rule.Metadata.Builder == nil {
		return "", false, nil
	}
	return rule.Metadata.Builder.Type, true, nil
}

func checkServerBuilderType(spaceID, ruleID, wantType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		gotType, present, err := serverBuilderType(spaceID, ruleID)
		if err != nil {
			return err
		}
		if !present || gotType != wantType {
			return fmt.Errorf("rule %s/%s metadata.builder.type = %q (present=%t), want %q", spaceID, ruleID, gotType, present, wantType)
		}
		return nil
	}
}

func checkServerBuilderAbsent(spaceID, ruleID string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		gotType, present, err := serverBuilderType(spaceID, ruleID)
		if err != nil {
			return err
		}
		if present {
			return fmt.Errorf("rule %s/%s still carries metadata.builder.type %q", spaceID, ruleID, gotType)
		}
		return nil
	}
}

// deleteRule removes a default-space rule outside Terraform.
func deleteRule(t *testing.T, ruleID string) {
	t.Helper()
	client, err := acceptanceKibanaClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client: %v", err)
	}
	if diags := kibanaoapi.DeleteAlertingV2Rule(context.Background(), client, "default", ruleID); diags.HasError() {
		t.Fatalf("failed to delete rule default/%s: %v", ruleID, diags)
	}
}

func checkResourceAbsent(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, found := s.RootModule().Resources[addr]; found {
			return fmt.Errorf("resource %s is still in state", addr)
		}
		return nil
	}
}

func checkAttrDiffers(addr, key string, previous *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var current string
		if err := rememberAttr(addr, key, &current)(s); err != nil {
			return err
		}
		if current == *previous {
			return fmt.Errorf("expected %s on %s to change from %q", key, addr, *previous)
		}
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
