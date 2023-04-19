package kibana_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceActionConnector(t *testing.T) {
	connectorName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceActionConnectorDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceActionConnectorCreate(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "name", connectorName),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "connector_type_id", ".index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_preconfigured", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedVersion),
				Config:   testAccResourceActionConnectorUpdate(connectorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "name", fmt.Sprintf("Updated %s", connectorName)),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "connector_type_id", ".index"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_deprecated", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_missing_secrets", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_action_connector.test_connector", "is_preconfigured", "false"),
				),
			},
		},
	})
}

func testAccResourceActionConnectorCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test_connector" {
  name         = "%s"
  config       = jsonencode({
	index             = ".kibana"
	refresh             = true
  })
  connector_type_id = ".index"
}
	`, name)
}

func testAccResourceActionConnectorUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_action_connector" "test_connector" {
  name         = "Updated %s"
  config       = jsonencode({
	index             = ".kibana"
	refresh             = false
  })
  connector_type_id = ".index"
}
	`, name)
}

func checkResourceActionConnectorDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_action_connector" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		rule, diags := kibana.GetActionConnector(context.Background(), client, compId.ResourceId, compId.ClusterId, ".index")
		if diags.HasError() {
			return fmt.Errorf("Failed to get action connector: %v", diags)
		}

		if rule != nil {
			return fmt.Errorf("Action connector (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
