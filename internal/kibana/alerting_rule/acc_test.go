package alerting_rule_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceAlertingRule(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))
	minSupportedFrequencyVersion := version.Must(version.NewSemver("8.7.0"))
	minSupportedAlertsFilterVersion := version.Must(version.NewSemver("8.9.0"))
	minSupportedAlertDelayVersion := version.Must(version.NewSemver("8.13.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	// Clean up any dangling rules from previous test runs
	cleanupDanglingRules := func() {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return // Ignore errors during cleanup
		}

		// Delete the rule IDs used in this test
		ruleIDs := []string{
			"bf33ce2d-9fc4-5131-a350-b5bd6482735c",
			"cf33ce2d-9fc4-5131-a350-b5bd6482736c",
		}
		for _, ruleID := range ruleIDs {
			kibana.DeleteAlertingRule(context.Background(), client, ruleID, "default")
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			cleanupDanglingRules()
		},
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}} 3","rule_id":"{{rule.id}} 1","rule_name":"{{rule.name}} 2"}]}`),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "alert_delay", "10"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleEnabledFalseOnCreate(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	// Clean up any dangling rules from previous test runs
	cleanupDanglingRules := func() {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return // Ignore errors during cleanup
		}
		kibana.DeleteAlertingRule(context.Background(), client, "df33ce2d-9fc4-5131-a350-b5bd6482737d", "default")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			cleanupDanglingRules()
		},
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

func TestAccResourceAlertingRuleFromSDK(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	// Clean up any dangling rules from previous test runs
	cleanupDanglingRules := func() {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return // Ignore errors during cleanup
		}
		kibana.DeleteAlertingRule(context.Background(), client, "ef33ce2d-9fc4-5131-a350-b5bd6482745e", "default")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			cleanupDanglingRules()
		},
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
				Config: testAccAlertingRuleSDKConfig(ruleName),
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
				Config:                   testAccAlertingRuleSDKConfig(ruleName),
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

func testAccAlertingRuleSDKConfig(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id      = "ef33ce2d-9fc4-5131-a350-b5bd6482745e"
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true

  params = jsonencode({
    "index" : [".test-index"],
    "timeField" : "@timestamp",
    "aggType" : "count",
    "groupBy" : "all",
    "timeWindowSize" : 5,
    "timeWindowUnit" : "m",
    "thresholdComparator" : ">",
    "threshold" : [1000]
  })
}
`, name)
}

func checkResourceAlertingRuleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_alerting_rule" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		rule, diags := kibana.GetAlertingRule(context.Background(), client, compId.ResourceId, compId.ClusterId)
		if diags.HasError() {
			return fmt.Errorf("Failed to get alerting rule: %v", diags)
		}

		if rule != nil {
			return fmt.Errorf("Alerting rule (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
