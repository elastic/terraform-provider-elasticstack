package kibana_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceKibanaConnector(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceConnector,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "name", "myconnector"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "space_id", "default"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "connector_type_id", ".slack"),
					resource.TestCheckResourceAttrSet("data.elasticstack_kibana_action_connector.myconnector", "connector_id"),
				),
			},
		},
	})
}

const testAccDataSourceConnector = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "slack" {
	name              = "myconnector"
	connector_type_id = ".slack"
	secrets = jsonencode({
	  webhookUrl = "https://internet.com"
	})
  }

data "elasticstack_kibana_action_connector" "myconnector" {
	name    = elasticstack_kibana_action_connector.slack.name
}
`
