package security_detection_rule_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceSecurityDetectionRule(t *testing.T) {
	resourceName := "elasticstack_kibana_security_detection_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccCheckSecurityDetectionRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityDetectionRuleConfig_basic("test-rule"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, "type", "query"),
					resource.TestCheckResourceAttr(resourceName, "query", "*:*"),
					resource.TestCheckResourceAttr(resourceName, "language", "kuery"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "50"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				Config: testAccSecurityDetectionRuleConfig_update("test-rule-updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "test-rule-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test security detection rule"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "risk_score", "75"),
				),
			},
		},
	})
}

func testAccCheckSecurityDetectionRuleDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	kbClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_detection_rule" {
			continue
		}

		// Parse ID to get space_id and rule_id
		parts := strings.Split(rs.Primary.ID, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid resource ID format: %s", rs.Primary.ID)
		}
		ruleId := parts[1]

		// Check if the rule still exists
		ruleObjectId := kbapi.SecurityDetectionsAPIRuleObjectId(uuid.MustParse(ruleId))
		params := &kbapi.ReadRuleParams{
			Id: &ruleObjectId,
		}

		response, err := kbClient.API.ReadRuleWithResponse(context.Background(), params)
		if err != nil {
			return fmt.Errorf("failed to read security detection rule: %v", err)
		}

		// If the rule still exists (status 200), it means destroy failed
		if response.StatusCode() == 200 {
			return fmt.Errorf("security detection rule (%s) still exists", ruleId)
		}

		// If we get a 404, that's expected - the rule was properly destroyed
		// Any other status code indicates an error
		if response.StatusCode() != 404 {
			return fmt.Errorf("unexpected status code when checking security detection rule: %d", response.StatusCode())
		}
	}

	return nil
}

func testAccSecurityDetectionRuleConfig_basic(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Test security detection rule"
  severity    = "medium"
  risk_score  = 50
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
}
`, name)
}

func testAccSecurityDetectionRuleConfig_update(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "%s"
  type        = "query"
  query       = "*:*"
  language    = "kuery"
  enabled     = true
  description = "Updated test security detection rule"
  severity    = "high"
  risk_score  = 75
  from        = "now-6m"
  to          = "now"
  interval    = "5m"
  author      = ["Test Author"]
  tags        = ["test", "automation"]
  license     = "Elastic License v2"
}
`, name)
}
