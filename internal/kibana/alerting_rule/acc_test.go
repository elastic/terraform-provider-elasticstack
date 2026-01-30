package alerting_rule_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var minAlertingRuleVersion = version.Must(version.NewVersion("8.0.0"))

func TestAccResourceAlertingRuleIndexThreshold(t *testing.T) {
	ruleName := "test-rule-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName := "test-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minAlertingRuleVersion),
				Config:   testAccResourceAlertingRuleIndexThreshold(ruleName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_alerting_rule.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "rule_type_id", ".index-threshold"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "interval", "1m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "enabled", "true"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minAlertingRuleVersion),
				Config:   testAccResourceAlertingRuleIndexThresholdUpdated(ruleName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_alerting_rule.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "name", ruleName+"-updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test", "interval", "5m"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleESQuery(t *testing.T) {
	ruleName := "test-esquery-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName := "test-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minAlertingRuleVersion),
				Config:   testAccResourceAlertingRuleESQuery(ruleName, indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_alerting_rule.test_esquery", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_esquery", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_esquery", "consumer", "alerts"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_esquery", "rule_type_id", ".es-query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_esquery", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleWithActions(t *testing.T) {
	ruleName := "test-actions-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName := "test-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	connectorName := "test-conn-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	// Actions with frequency require 8.6+
	minVersionWithFrequency := version.Must(version.NewVersion("8.6.0"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionWithFrequency),
				Config:   testAccResourceAlertingRuleWithActions(ruleName, indexName, connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_alerting_rule.test_actions", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_actions", "name", ruleName),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_actions", "actions.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceAlertingRuleMalformedParams(t *testing.T) {
	ruleName := "test-malformed-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(minAlertingRuleVersion),
				Config:      testAccResourceAlertingRuleMalformedParams(ruleName),
				ExpectError: regexp.MustCompile(`(?i)(invalid|error|bad request|failed)`),
			},
		},
	})
}

func TestAccResourceAlertingRuleInvalidJSON(t *testing.T) {
	ruleName := "test-invalidjson-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:    versionutils.CheckIfVersionIsUnsupported(minAlertingRuleVersion),
				Config:      testAccResourceAlertingRuleInvalidJSON(ruleName),
				ExpectError: regexp.MustCompile(`(?i)(invalid|json|syntax|unexpected)`),
			},
		},
	})
}

func testAccResourceAlertingRuleIndexThreshold(ruleName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_alerting_rule" "test" {
	name         = "%s"
	consumer     = "alerts"
	rule_type_id = ".index-threshold"
	interval     = "1m"
	enabled      = true

	params = jsonencode({
		aggType             = "count"
		thresholdComparator = ">="
		timeWindowSize      = 5
		timeWindowUnit      = "m"
		groupBy             = "all"
		threshold           = [100]
		index               = [elasticstack_elasticsearch_index.test_index.name]
		timeField           = "@timestamp"
	})
}
`, indexName, ruleName)
}

func testAccResourceAlertingRuleIndexThresholdUpdated(ruleName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "test_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_alerting_rule" "test" {
	name         = "%s-updated"
	consumer     = "alerts"
	rule_type_id = ".index-threshold"
	interval     = "5m"
	enabled      = true

	params = jsonencode({
		aggType             = "count"
		thresholdComparator = ">="
		timeWindowSize      = 10
		timeWindowUnit      = "m"
		groupBy             = "all"
		threshold           = [50]
		index               = [elasticstack_elasticsearch_index.test_index.name]
		timeField           = "@timestamp"
	})
}
`, indexName, ruleName)
}

func testAccResourceAlertingRuleESQuery(ruleName, indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "test_index_esquery" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_alerting_rule" "test_esquery" {
	name         = "%s"
	consumer     = "alerts"
	rule_type_id = ".es-query"
	interval     = "1m"
	enabled      = true

	params = jsonencode({
		searchType               = "esQuery"
		timeWindowSize           = 5
		timeWindowUnit           = "m"
		threshold                = [10]
		thresholdComparator      = ">"
		size                     = 100
		aggType                  = "count"
		groupBy                  = "all"
		excludeHitsFromPreviousRun = true
		esQuery                  = jsonencode({
			query = {
				match_all = {}
			}
		})
		index                    = [elasticstack_elasticsearch_index.test_index_esquery.name]
		timeField                = "@timestamp"
	})
}
`, indexName, ruleName)
}

func testAccResourceAlertingRuleWithActions(ruleName, indexName, connectorName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "test_index_actions" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_action_connector" "test_connector" {
	name              = "%s"
	connector_type_id = ".server-log"
}

resource "elasticstack_kibana_alerting_rule" "test_actions" {
	name         = "%s"
	consumer     = "alerts"
	rule_type_id = ".index-threshold"
	interval     = "1m"
	enabled      = true

	params = jsonencode({
		aggType             = "count"
		thresholdComparator = ">"
		timeWindowSize      = 5
		timeWindowUnit      = "m"
		groupBy             = "all"
		threshold           = [100]
		index               = [elasticstack_elasticsearch_index.test_index_actions.name]
		timeField           = "@timestamp"
	})

	actions {
		id    = elasticstack_kibana_action_connector.test_connector.connector_id
		group = "threshold met"
		params = jsonencode({
			level   = "info"
			message = "Alert triggered: {{context.message}}"
		})
		frequency {
			summary     = false
			notify_when = "onActiveAlert"
		}
	}
}
`, indexName, connectorName, ruleName)
}

func testAccResourceAlertingRuleMalformedParams(ruleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_malformed" {
	name         = "%s"
	consumer     = "alerts"
	rule_type_id = ".index-threshold"
	interval     = "1m"
	enabled      = true

	params = jsonencode({
		# Missing required fields like index, threshold, etc.
		aggType = "count"
	})
}
`, ruleName)
}

func testAccResourceAlertingRuleInvalidJSON(ruleName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_kibana_alerting_rule" "test_invalid_json" {
	name         = "%s"
	consumer     = "alerts"
	rule_type_id = ".index-threshold"
	interval     = "1m"
	enabled      = true

	# Invalid JSON - missing closing brace
	params = "{invalid json"
}
`, ruleName)
}
