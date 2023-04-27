package fleet_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOutput(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOutputCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "ssl.verification_mode: none"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				Config: testAccResourceOutputUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "ssl.verification_mode: none"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
		},
	})
}

func testAccResourceOutputCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "elasticsearch"
  config_yaml          = "ssl.verification_mode: none"
  default_integrations = false
  default_monitoring   = false
  hosts                = [
    "https://elasticsearch:9200"
  ]
}
`, fmt.Sprintf("Output %s", id))
}

func testAccResourceOutputUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "elasticsearch"
  config_yaml          = "ssl.verification_mode: none"
  default_integrations = false
  default_monitoring   = false
  hosts                = [
    "https://elasticsearch:9200"
  ]
}

`, fmt.Sprintf("Updated Output %s", id))
}

func checkResourceOutputDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_output" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		packagePolicy, diag := fleet.ReadOutput(context.Background(), fleetClient, rs.Primary.ID)
		if diag.HasError() {
			return fmt.Errorf(diag[0].Summary)
		}
		if packagePolicy != nil {
			return fmt.Errorf("output id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}