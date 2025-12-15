package kibana_test

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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac5471624b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
				),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac5471624b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1d"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.0", "first"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.1", "second"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac5471624b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "10m"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac5471624b"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.0", "first"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.1", "second"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}} 3","rule_id":"{{rule.id}} 1","rule_name":"{{rule.name}} 2"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActiveAlert"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "2h"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac54716255"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", "logs.alert.document.count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "logs.threshold.fired"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.kql", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.days.0", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.days.1", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.days.2", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.timezone", "Africa/Accra"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.hours_start", "01:00"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.0.hours_end", "07:00"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac54716255"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", "logs.alert.document.count"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "logs.threshold.fired"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.kql", `kibana.alert.action_group: "slo.burnRate.alert"`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.alerts_filter.0.timeframe.#", "0"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac54716255"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "10m"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac54716255"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "notify_when", ""),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.group", "threshold met"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.params", `{"documents":[{"message":"{{context.message}}","rule_id":"{{rule.id}}","rule_name":"{{rule.name}}"}]}`),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.summary", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.notify_when", "onActionGroupChange"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "actions.0.frequency.0.throttle", "10m"),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "rule_id", "af22bd1c-8fb3-4020-9249-a4ac5471624c"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule_disabled", "enabled", "false"),
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
