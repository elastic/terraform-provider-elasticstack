package fleet_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: Determine actual min version
var minVersionAgentPolicy = version.Must(version.NewVersion("8.7.0"))

func TestAccResourceAgentPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
				),
			},
		},
	})
}

func testAccResourceAgentPolicyCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy"
  monitor_logs = true
  monitor_metrics = true
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id))
}

func testAccResourceAgentPolicyUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  fleet {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "This policy was updated"
  monitor_logs = true
  monitor_metrics = true
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id))
}

func checkResourceAgentPolicyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_fleet_agent_policy" {
			continue
		}

		fleetClient, err := client.GetFleetClient()
		if err != nil {
			return err
		}
		packagePolicy, diag := fleet.ReadAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
		if diag.HasError() {
			return fmt.Errorf(diag[0].Summary)
		}
		if packagePolicy != nil {
			return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
