package detection_rule_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceDetectionRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDetectionRule_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "name", "Test Detection Rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "description", "A test detection rule"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "type", "query"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "severity", "medium"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "risk_score", "50"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_detection_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_detection_rule.test", "rule_id"),
				),
			},
		},
	})
}

const testAccResourceDetectionRule_basic = `
resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = "Test Detection Rule"
  description = "A test detection rule"
  type        = "query"
  severity    = "medium"
  risk_score  = 50
  enabled     = true
  query       = "user.name:*"
  language    = "kuery"
  
  tags = ["test", "terraform"]
}
`
