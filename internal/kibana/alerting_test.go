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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAlertingRule(t *testing.T) {
	minSupportedVersion := version.Must(version.NewSemver("7.14.0"))
	minSupportedFrequencyVersion := version.Must(version.NewSemver("8.7.0"))
	minSupportedAlertsFilterVersion := version.Must(version.NewSemver("8.9.0"))
	minSupportedAlertDelayVersion := version.Must(version.NewSemver("8.13.0"))

	t.Setenv("KIBANA_API_KEY", "")

	ruleName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAlertingRuleDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceAlertingRuleCreate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceAlertingRuleUpdate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedFrequencyVersion),
				Config:   testAccResourceAlertingRuleWithFrequencyCreate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedFrequencyVersion),
				Config:   testAccResourceAlertingRuleWithFrequencyUpdate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedAlertsFilterVersion),
				Config:   testAccResourceAlertingRuleWithAlertsFilterCreate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedAlertsFilterVersion),
				Config:   testAccResourceAlertingRuleWithAlertsFilterUpdate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedAlertDelayVersion),
				Config:   testAccResourceAlertingRuleWithAlertDelayCreate(ruleName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedAlertDelayVersion),
				Config:   testAccResourceAlertingRuleWithAlertDelayUpdate(ruleName),
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

func testAccResourceAlertingRuleCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac5471624b"
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
}
	`, name)
}

func testAccResourceAlertingRuleUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "Updated %s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac5471624b"
  consumer     = "alerts"
  notify_when  = "onActiveAlert"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1d"
  enabled      = false
  tags         = ["first", "second"]
}
	`, name)
}

// Frequency - v8.6.0

func testAccResourceAlertingRuleWithFrequencyCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac5471624b"
  consumer     = "alerts"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

	frequency {
	  summary     = true
	  notify_when = "onActionGroupChange"
      throttle    = "10m"
	}
  }
}
	`, name)
}

func testAccResourceAlertingRuleWithFrequencyUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
	name              = "my_index_connector"
	connector_type_id = ".index"
	config = jsonencode({
	  index              = "my-index"
	  executionTimeField = "alert_date"
	})
  }

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "Updated %s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac5471624b"
  consumer     = "alerts"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "10m"
  enabled      = false
  tags         = ["first", "second"]

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}} 1",
        "rule_name" : "{{rule.name}} 2",
        "message" : "{{context.message}} 3"
      }]
    })

	frequency {
	  summary     = false
	  notify_when = "onActiveAlert"
      throttle    = "2h"
	}
  }
}
	`, name)
}

// Alerts Filter - v8.9.0

func testAccResourceAlertingRuleWithAlertsFilterCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac54716255"
  consumer     = "alerts"
  params       = jsonencode({
    "timeSize": 5,
    "timeUnit": "m",
    "logView": {
      "type": "log-view-reference",
      "logViewId": "default"
    },
    "count": {
      "value": 75,
      "comparator": "more than"
    },
    "criteria": [
      {
        "field": "_id",
        "comparator": "matches",
        "value": "33"
      }
    ]
  })
  rule_type_id = "logs.alert.document.count"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "logs.threshold.fired"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = true
      notify_when = "onActionGroupChange"
        throttle    = "10m"
    }

    alerts_filter {
      timeframe {
        days        = [1,2,3]
        timezone    = "Africa/Accra"
        hours_start = "01:00"
        hours_end   = "07:00"
      }
    }
  }
}
	`, name)
}

func testAccResourceAlertingRuleWithAlertsFilterUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac54716255"
  consumer     = "alerts"
  params       = jsonencode({
    "timeSize": 5,
    "timeUnit": "m",
    "logView": {
      "type": "log-view-reference",
      "logViewId": "default"
    },
    "count": {
      "value": 75,
      "comparator": "more than"
    },
    "criteria": [
      {
        "field": "_id",
        "comparator": "matches",
        "value": "33"
      }
    ]
  })
  rule_type_id = "logs.alert.document.count"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "logs.threshold.fired"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = true
      notify_when = "onActionGroupChange"
        throttle    = "10m"
    }

    alerts_filter {
      kql = "kibana.alert.action_group: \"slo.burnRate.alert\""
    }
  }
}
	`, name)
}

// Alert Delay - v8.13.0

func testAccResourceAlertingRuleWithAlertDelayCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac54716255"
  consumer     = "alerts"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = true
      notify_when = "onActionGroupChange"
        throttle    = "10m"
    }
  }

  alert_delay  = 4
}
	`, name)
}

func testAccResourceAlertingRuleWithAlertDelayUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = "my-index"
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "test_rule" {
  name         = "%s"
  rule_id 	   = "af22bd1c-8fb3-4020-9249-a4ac54716255"
  consumer     = "alerts"
  params       = jsonencode({
	aggType             = "avg"
	groupBy             = "top"
	termSize            = 10
	timeWindowSize      = 10
	timeWindowUnit      = "s"
	threshold           = [10]
	thresholdComparator = ">"
	index               = ["test-index"]
	timeField           = "@timestamp"
	aggField            = "version"
	termField           = "name"
  })
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })

    frequency {
      summary     = true
      notify_when = "onActionGroupChange"
        throttle    = "10m"
    }
  }

  alert_delay  = 10
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
