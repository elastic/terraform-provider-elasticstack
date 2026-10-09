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

package alertingrules_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const alertingRulesDataSourceAddr = "data.elasticstack_kibana_alerting_rules.test"

func TestMain(m *testing.M) {
	originalAPIKey, hadAPIKey := os.LookupEnv("KIBANA_API_KEY")
	if err := os.Setenv("KIBANA_API_KEY", ""); err != nil {
		panic(err)
	}

	exitCode := m.Run()

	if hadAPIKey {
		if err := os.Setenv("KIBANA_API_KEY", originalAPIKey); err != nil {
			panic(err)
		}
	} else if err := os.Unsetenv("KIBANA_API_KEY"); err != nil {
		panic(err)
	}

	os.Exit(exitCode)
}

func TestAccDataSourceKibanaAlertingRules_lookup(t *testing.T) {
	ruleID := uuid.New().String()
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(name),
					"rule_id": config.StringVariable(ruleID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "space_id", "default"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "id", "default/"+ruleID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "1"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.id", ruleID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.name", name),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.consumer", "alerts"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_unknownRuleID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"rule_id": config.StringVariable(uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_unknownSpace(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable("missing-" + sdkacctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")),
					"rule_id":  config.StringVariable(uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`(?i)(forbidden|not found|unexpected status|HTTP 40)`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_compositeRuleID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"rule_id": config.StringVariable("other-space/" + uuid.New().String()),
				},
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_filter(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(12, "abcdefghijklmnopqrstuvwxyz0123456789")
	enabledID := uuid.New().String()
	disabledID := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id":         config.StringVariable(spaceID),
					"enabled_rule_id":  config.StringVariable(enabledID),
					"disabled_rule_id": config.StringVariable(disabledID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "1"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.id", enabledID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_allRules(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(12, "abcdefghijklmnopqrstuvwxyz0123456789")
	firstID := uuid.New().String()
	secondID := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id":  config.StringVariable(spaceID),
					"first_id":  config.StringVariable(firstID),
					"second_id": config.StringVariable(secondID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "id", spaceID),
					checkRuleIDSet(alertingRulesDataSourceAddr, firstID, secondID),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_orderedByName(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(12, "abcdefghijklmnopqrstuvwxyz0123456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"a_id":     config.StringVariable(uuid.New().String()),
					"b_id":     config.StringVariable(uuid.New().String()),
					"c_id":     config.StringVariable(uuid.New().String()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "3"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.name", "a-rule"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.1.name", "b-rule"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.2.name", "c-rule"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_noMatches(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(12, "abcdefghijklmnopqrstuvwxyz0123456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"rule_id":  config.StringVariable(uuid.New().String()),
					"name":     config.StringVariable("present-rule"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_emptySpace(t *testing.T) {
	spaceID := sdkacctest.RandStringFromCharSet(12, "abcdefghijklmnopqrstuvwxyz0123456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "space_id", spaceID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "id", spaceID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_defaultSpaceSearch(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "space_id", "default"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "id", "default"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_readDoesNotChangeRule(t *testing.T) {
	ruleID := uuid.New().String()
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(name),
					"rule_id": config.StringVariable(ruleID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "name", name),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "enabled", "true"),
					checkAlertingRuleUnchanged("default", ruleID, name, true),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_invalidFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ExpectError:              regexp.MustCompile(`(?i)filter`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_conflictingFilters(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ExpectError:              regexp.MustCompile(`(?i)cannot be specified when`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_emptyRuleID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ExpectError:              regexp.MustCompile(`(?i)string length must be at least`),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_tags(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	taggedID := uuid.New().String()
	untaggedID := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":             config.StringVariable(name),
					"tagged_rule_id":   config.StringVariable(taggedID),
					"untagged_rule_id": config.StringVariable(untaggedID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_alerting_rules.tagged", "rules.0.tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_alerting_rules.tagged", "rules.0.tags.*", "tag-one"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_kibana_alerting_rules.tagged", "rules.0.tags.*", "tag-two"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_alerting_rules.untagged", "rules.0.tags.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceKibanaAlertingRules_scheduledTaskID(t *testing.T) {
	ruleID := uuid.New().String()
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(name),
					"rule_id": config.StringVariable(ruleID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(alertingRulesDataSourceAddr, "rules.0.scheduled_task_id"),
				),
			},
		},
	})
}

// TestAccDataSourceKibanaAlertingRules_executionStatus confirms last_execution_status and
// last_execution_date surface once Kibana has actually run the rule; immediately after
// creation both are null (see the resource-level equivalent in alertingrule/acc_test.go).
func TestAccDataSourceKibanaAlertingRules_executionStatus(t *testing.T) {
	ruleID := uuid.New().String()
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)
	configVars := config.Variables{
		"name":    config.StringVariable(name),
		"rule_id": config.StringVariable(ruleID),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables:          configVars,
			},
			{
				PreConfig: func() {
					waitForAlertingRuleFirstExecution(t, clients.DefaultSpaceID, ruleID)
				},
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables:          configVars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(alertingRulesDataSourceAddr, "rules.0.last_execution_status"),
					resource.TestCheckResourceAttrSet(alertingRulesDataSourceAddr, "rules.0.last_execution_date"),
				),
			},
		},
	})
}

func waitForAlertingRuleFirstExecution(t *testing.T, spaceID, ruleID string) {
	t.Helper()

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("kibana client: %v", err)
	}
	oapi := client.GetKibanaOapiClient()
	ctx := context.Background()

	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		rule, diags := kibanaoapi.GetAlertingRule(ctx, oapi, spaceID, ruleID)
		if diags.HasError() || rule == nil {
			t.Fatalf("get rule: %v", diags)
		}
		if rule.ExecutionStatus.LastExecutionDate != nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Kibana did not execute the rule in time")
}

// TestAccDataSourceKibanaAlertingRules_ruleIDWithSlash confirms rule_id is always fetched
// literally rather than parsed as a "<space>/<rule_id>" composite: a rule whose actual ID
// contains a slash is still found via a plain rule_id lookup in its own space. See
// TestRead_compositeRuleIDIsFetchedLiterally in datasource_read_test.go for the unit-level
// equivalent, and TestAccDataSourceKibanaAlertingRules_compositeRuleID for the not-found case.
func TestAccDataSourceKibanaAlertingRules_ruleIDWithSlash(t *testing.T) {
	ruleID := "sub/" + uuid.New().String()
	name := sdkacctest.RandStringFromCharSet(12, sdkacctest.CharSetAlphaNum)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("read"),
				ConfigVariables: config.Variables{
					"name":    config.StringVariable(name),
					"rule_id": config.StringVariable(ruleID),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.#", "1"),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.id", ruleID),
					resource.TestCheckResourceAttr(alertingRulesDataSourceAddr, "rules.0.name", name),
				),
			},
		},
	})
}

var checkAlertingRuleDestroy = checks.KibanaResourceDestroyCheckCompositeID(
	"elasticstack_kibana_alerting_rule",
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, ruleID string) (bool, error) {
		rule, diags := kibanaoapi.GetAlertingRule(ctx, client, spaceID, ruleID)
		if diags.HasError() {
			return false, fmt.Errorf("failed to get alerting rule: %v", diags)
		}
		return rule != nil, nil
	},
)

func checkRuleIDSet(addr string, ids ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("data source %q not found in state", addr)
		}
		got := map[string]struct{}{}
		for key, value := range rs.Primary.Attributes {
			if len(key) > 3 && key[len(key)-3:] == ".id" && key != "id" {
				got[value] = struct{}{}
			}
		}
		if len(got) != len(ids) {
			return fmt.Errorf("expected %d rules, got %d (%v)", len(ids), len(got), got)
		}
		for _, id := range ids {
			if _, ok := got[id]; !ok {
				return fmt.Errorf("rules missing id %q (got %v)", id, got)
			}
		}
		return nil
	}
}

func checkAlertingRuleUnchanged(spaceID, ruleID, name string, enabled bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		rule, diags := kibanaoapi.GetAlertingRule(context.Background(), client.GetKibanaOapiClient(), spaceID, ruleID)
		if diags.HasError() {
			return fmt.Errorf("failed to get alerting rule: %v", diags)
		}
		if rule == nil {
			return fmt.Errorf("alerting rule %q was deleted by the data source read", ruleID)
		}
		if rule.Name != name {
			return fmt.Errorf("rule name changed from %q to %q", name, rule.Name)
		}
		if rule.Enabled == nil || *rule.Enabled != enabled {
			return fmt.Errorf("rule enabled changed, got %v", rule.Enabled)
		}
		return nil
	}
}
