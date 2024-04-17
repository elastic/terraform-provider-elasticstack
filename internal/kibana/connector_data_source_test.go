package kibana_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "space_id", "supdawg"),
					resource.TestCheckResourceAttr("data.elasticstack_kibana_action_connector.myconnector", "connector_type_id", ".slack"),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "elasticsearch.0.run_as", []string{"elastic", "kibana"}),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.base", []string{"all"}),
					utils.TestCheckResourceListAttr("data.elasticstack_kibana_security_role.test", "kibana.0.spaces", []string{"default"}),
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
	space_id		  = "supdawg"
	connector_type_id = ".slack"
	secrets = jsonencode({
	  webhookUrl = "https://internet.com"
	})
  }

data "elasticstack_kibana_action_connector" "myconnector" {
	name    = "myconnector"
	space_id = "supdawg"
}

`
