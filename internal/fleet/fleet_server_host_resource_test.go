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

func TestAccResourceFleetServerHost(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceFleetServerHostDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFleetServerHostCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
			{
				Config: testAccResourceFleetServerHostUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "name", fmt.Sprintf("Updated FleetServerHost %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "default", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_server_host.test_host", "hosts.0", "https://fleet-server:8220"),
				),
			},
		},
	})
}

func testAccResourceFleetServerHostCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name               = "%s"
  default            =  false
  hosts              = [
    "https://fleet-server:8220"
  ]
}
`, fmt.Sprintf("FleetServerHost %s", id))
}

func testAccResourceFleetServerHostUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_server_host" "test_host" {
  name               = "%s"
  default            =  false
  hosts              = [
    "https://fleet-server:8220"
  ]
}

`, fmt.Sprintf("Updated FleetServerHost %s", id))
}

func checkResourceFleetServerHostDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_server_host" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		packagePolicy, diag := fleet.ReadFleetServerHost(context.Background(), fleetClient, rs.Primary.ID)
		if diag.HasError() {
			return fmt.Errorf(diag[0].Summary)
		}
		if packagePolicy != nil {
			return fmt.Errorf("FleetServerHost id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
