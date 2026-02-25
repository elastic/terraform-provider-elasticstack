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
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	alertingRuleActionParamsDefault = `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`
	alertingRuleActionParamsUpdated = `{"documents":[{"message":"{{context.message}} 3","rule_id":"{{rule.id}} 1","rule_name":"{{rule.name}} 2"}]}`
)

func TestAccResourceAlertingRule(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))
	minSupportedFrequencyVersion := version.Must(version.NewSemver("8.7.0"))
	minSupportedAlertsFilterVersion := version.Must(version.NewSemver("8.9.0"))
	minSupportedAlertDelayVersion := version.Must(version.NewSemver("8.13.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "bf33ce2d-9fc4-5131-a350-b5bd6482735c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
				),
			},
			// ImportState testing
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ResourceName:             "elasticstack_kibana_alerting_rule.test_rule",
				ImportState:              true,
				ImportStateVerify:        true,
				// notify_when may not be returned by the API in newer versions where it's deprecated
				// last_execution_date and last_execution_status change as Kibana executes the rule
				ImportStateVerifyIgnore: []string{"notify_when", "last_execution_date", "last_execution_status"},
				ConfigDirectory:         acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Updated %s", ruleName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", fmt.Sprintf("Updated %s", ruleName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "bf33ce2d-9fc4-5131-a350-b5bd6482735c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "false"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.*", "first"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.*", "second"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedFrequencyVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("frequency_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "bf33ce2d-9fc4-5131-a350-b5bd6482735c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedFrequencyVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("frequency_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(fmt.Sprintf("Updated %s", ruleName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", fmt.Sprintf("Updated %s", ruleName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "bf33ce2d-9fc4-5131-a350-b5bd6482735c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "false"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.*", "first"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.*", "second"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsUpdated,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "2h"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAlertsFilterVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("alerts_filter_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "cf33ce2d-9fc4-5131-a350-b5bd6482736c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", "logs.alert.document.count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "logs.threshold.fired"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.days.0", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.days.1", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.days.2", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.timezone", "Africa/Accra"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.hours_start", "01:00"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.timeframe.hours_end", "07:00"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAlertsFilterVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("alerts_filter_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "cf33ce2d-9fc4-5131-a350-b5bd6482736c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", "logs.alert.document.count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "logs.threshold.fired"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.kql", `kibana.alert.action_group: "slo.burnRate.alert"`),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAlertDelayVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("alert_delay_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "cf33ce2d-9fc4-5131-a350-b5bd6482736c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "alert_delay", "4"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedAlertDelayVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("alert_delay_update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "cf33ce2d-9fc4-5131-a350-b5bd6482736c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "alert_delay", "10"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("actions_no_frequency_create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af33ce2d-9fc4-5131-a350-b5bd6482746f"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr(
						"elasticstack_kibana_alerting_rule.test_rule",
						"actions.0.params",
						alertingRuleActionParamsDefault,
					),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.1.group", "recovered"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.1.params", `{"documents":[{"message":"Recovered","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleParamsLifecycle(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_explicit"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "ff33ce2d-9fc4-5131-a350-b5bd6482799f"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "aggType", "avg"),
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "aggField", "version"),
					// Kibana injects groupBy="all" even when config omits it.
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "groupBy", "all"),
					testCheckAlertingRuleStateParamsMissingKeys("elasticstack_kibana_alerting_rule.test_rule", "groupBy"),
					testCheckAlertingRuleStateParamsHasKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"aggType",
						"aggField",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
					testCheckAlertingRuleStateParamsOnlyKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"aggType",
						"aggField",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("remove_aggtype"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					// When aggType/aggField are removed from config, Kibana should revert to its defaults.
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "aggType", "count"),
					testCheckAlertingRuleAPIParamAbsentOrEmpty("elasticstack_kibana_alerting_rule.test_rule", "aggField"),
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "groupBy", "all"),
					testCheckAlertingRuleStateParamsMissingKeys("elasticstack_kibana_alerting_rule.test_rule", "aggType", "aggField", "groupBy"),
					testCheckAlertingRuleStateParamsHasKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
					testCheckAlertingRuleStateParamsOnlyKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("add_aggtype"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "aggType", "avg"),
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "aggField", "version"),
					testCheckAlertingRuleAPIParamStringEquals("elasticstack_kibana_alerting_rule.test_rule", "groupBy", "all"),
					testCheckAlertingRuleStateParamsMissingKeys("elasticstack_kibana_alerting_rule.test_rule", "groupBy"),
					testCheckAlertingRuleStateParamsHasKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"aggType",
						"aggField",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
					testCheckAlertingRuleStateParamsOnlyKeys(
						"elasticstack_kibana_alerting_rule.test_rule",
						"aggType",
						"aggField",
						"timeWindowSize",
						"timeWindowUnit",
						"threshold",
						"thresholdComparator",
						"index",
						"timeField",
					),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleEnabledFalseOnCreate(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "rule_id", "df33ce2d-9fc4-5131-a350-b5bd6482737d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleInconsistentParams(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.13.0"))

	t.Setenv("KIBANA_API_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inconsistent_params"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "name", "[Motel Services] Kafka Error Rate"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "consumer", "infrastructure"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "rule_type_id", ".es-query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "alert_delay", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "actions.#", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("inconsistent_params"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.kafka_error_alert", "actions.#", "2"),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceAlertingRuleFromSDK/create/rule.tf
var sdkCreateTestConfig string

func TestAccResourceAlertingRuleFromSDK(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				// Create the alerting rule with the last provider version where it was built on the SDK
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.13.1",
					},
				},
				Config: sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "ef33ce2d-9fc4-5131-a350-b5bd6482745e"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
				),
			},
			{
				// Upgrade to current PFW provider
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "ef33ce2d-9fc4-5131-a350-b5bd6482745e"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleAlertDelay(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("8.13.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAlertingRuleDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "rule_type_id", ".es-query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "enabled", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "tags.*", "autoops"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "alert_delay", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(ruleName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "rule_type_id", ".es-query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "enabled", "true"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "tags.*", "autoops"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.autoops_service_crashloopbackoff", "alert_delay", "1"),
				),
			},
		},
	})
}

func checkResourceAlertingRuleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_alerting_rule" {
			continue
		}
		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		rule, diags := kibanaoapi.GetAlertingRule(context.Background(), oapiClient, compID.ClusterID, compID.ResourceID)
		if diags.HasError() {
			return fmt.Errorf("Failed to get alerting rule: %v", diags)
		}

		if rule != nil {
			return fmt.Errorf("Alerting rule (%s) still exists", compID.ResourceID)
		}
	}
	return nil
}

func testCheckAlertingRuleAPIParams(resourceName string, check func(params map[string]any) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %q not found in state", resourceName)
		}

		compID, _ := clients.CompositeIDFromStr(rs.Primary.ID)

		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		oapiClient, err := client.GetKibanaOapiClient()
		if err != nil {
			return err
		}

		rule, diags := kibanaoapi.GetAlertingRule(context.Background(), oapiClient, compID.ClusterID, compID.ResourceID)
		if diags.HasError() {
			return fmt.Errorf("failed to get alerting rule: %v", diags)
		}
		if rule == nil {
			return fmt.Errorf("alerting rule (%s) not found", compID.ResourceID)
		}

		params := rule.Params
		if params == nil {
			params = map[string]any{}
		}
		return check(params)
	}
}

//nolint:unparam
func testCheckAlertingRuleAPIParamStringEquals(resourceName, key, expected string) resource.TestCheckFunc {
	return testCheckAlertingRuleAPIParams(resourceName, func(params map[string]any) error {
		v, ok := params[key]
		if !ok {
			return fmt.Errorf("expected Kibana params to contain %q, but it was absent (params=%v)", key, params)
		}
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected Kibana params %q to be a string, got %T (%v)", key, v, v)
		}
		if s != expected {
			return fmt.Errorf("expected Kibana params %q to equal %q, got %q", key, expected, s)
		}
		return nil
	})
}

func testCheckAlertingRuleAPIParamAbsentOrEmpty(resourceName, key string) resource.TestCheckFunc {
	return testCheckAlertingRuleAPIParams(resourceName, func(params map[string]any) error {
		v, ok := params[key]
		if !ok {
			return nil
		}
		if s, ok := v.(string); ok && s == "" {
			return nil
		}
		return fmt.Errorf("expected Kibana params %q to be absent (or empty string), got %T (%v)", key, v, v)
	})
}

func testCheckAlertingRuleStateParamsMissingKeys(resourceName string, keys ...string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(resourceName, "params", func(value string) error {
		var params map[string]any
		if err := json.Unmarshal([]byte(value), &params); err != nil {
			return fmt.Errorf("failed to unmarshal Terraform state params: %w", err)
		}
		for _, key := range keys {
			if _, exists := params[key]; exists {
				return fmt.Errorf("expected Terraform state params to omit key %q (API-injected default), but it was present (params=%v)", key, params)
			}
		}
		return nil
	})
}

func testCheckAlertingRuleStateParamsHasKeys(resourceName string, keys ...string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(resourceName, "params", func(value string) error {
		var params map[string]any
		if err := json.Unmarshal([]byte(value), &params); err != nil {
			return fmt.Errorf("failed to unmarshal Terraform state params: %w", err)
		}
		for _, key := range keys {
			if _, exists := params[key]; !exists {
				return fmt.Errorf("expected Terraform state params to contain key %q, but it was absent (params=%v)", key, params)
			}
		}
		return nil
	})
}

func testCheckAlertingRuleStateParamsOnlyKeys(resourceName string, allowedKeys ...string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(resourceName, "params", func(value string) error {
		var params map[string]any
		if err := json.Unmarshal([]byte(value), &params); err != nil {
			return fmt.Errorf("failed to unmarshal Terraform state params: %w", err)
		}

		allowed := make(map[string]struct{}, len(allowedKeys))
		for _, k := range allowedKeys {
			allowed[k] = struct{}{}
		}

		for k := range params {
			if _, ok := allowed[k]; !ok {
				return fmt.Errorf("expected Terraform state params to contain only keys %v, but found unexpected key %q (params=%v)", allowedKeys, k, params)
			}
		}
		return nil
	})
}
