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
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "interval", "10m"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "enabled", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.0", "first"),
					resource.TestCheckResourceAttr("elasticstack_kibana_alerting_rule.test_rule", "tags.1", "second"),
				),
			},
		},
	})
}

func testAccResourceAlertingRuleCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
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
  elasticsearch {}
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
  interval     = "10m"
  enabled      = false
  tags         = ["first", "second"]
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
