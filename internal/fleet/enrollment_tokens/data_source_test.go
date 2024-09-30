package enrollment_tokens_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var minVersionEnrollmentTokens = version.Must(version.NewVersion("8.6.0"))

func TestAccDataSourceEnrollmentTokens(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnrollmentTokens),
				Config:   testAccDataSourceEnrollmentToken,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "policy_id", "223b1bf8-240f-463f-8466-5062670d0754"),
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "tokens.0.policy_id", "223b1bf8-240f-463f-8466-5062670d0754"),
				),
			},
		},
	})
}

const testAccDataSourceEnrollmentToken = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test" {
  policy_id   = "223b1bf8-240f-463f-8466-5062670d0754"
  name        = "Test Agent Policy"
  namespace   = "default"
  description = "Agent Policy for testing Enrollment Tokens"
}

data "elasticstack_fleet_enrollment_tokens" "test" {
  policy_id = elasticstack_fleet_agent_policy.test.policy_id
}
`

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
		policy, diags := fleet.ReadAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
		if diags.HasError() {
			return utils.FwDiagsAsError(diags)
		}
		if policy != nil {
			return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
