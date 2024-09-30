package agent_policy_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var minVersionAgentPolicy = version.Must(version.NewVersion("8.6.0"))

func TestAccResourceAgentPolicyFromSDK(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceAgentPolicyDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:                   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
		},
	})
}

func TestAccResourceAgentPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicyDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyUpdate(policyName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Updated Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "This policy was updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "false"),
				),
			},
			{
				SkipFunc:                versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:                  testAccResourceAgentPolicyUpdate(policyName, false),
				ResourceName:            "elasticstack_fleet_agent_policy.test_policy",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func TestAccResourceAgentPolicySkipDestroy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceAgentPolicySkipDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionAgentPolicy),
				Config:   testAccResourceAgentPolicyCreate(policyName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "name", fmt.Sprintf("Policy %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "namespace", "default"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "description", "Test Agent Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_logs", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "monitor_metrics", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_agent_policy.test_policy", "skip_destroy", "true"),
				),
			},
		},
	})
}

func testAccResourceAgentPolicyCreate(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "Test Agent Policy"
  monitor_logs = true
  monitor_metrics = false
  skip_destroy = %t
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

`, fmt.Sprintf("Policy %s", id), skipDestroy)
}

func testAccResourceAgentPolicyUpdate(id string, skipDestroy bool) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name        = "%s"
  namespace   = "default"
  description = "This policy was updated"
  monitor_logs = false
  monitor_metrics = true
  skip_destroy = %t
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}
`, fmt.Sprintf("Updated Policy %s", id), skipDestroy)
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

func checkResourceAgentPolicySkipDestroy(s *terraform.State) error {
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
		if policy == nil {
			return fmt.Errorf("agent policy id=%v does not exist, but should still exist when skip_destroy is true", rs.Primary.ID)
		}

		if diags = fleet.DeleteAgentPolicy(context.Background(), fleetClient, rs.Primary.ID); diags.HasError() {
			return errors.New(diags.Errors()[0].Summary())
		}
	}
	return nil
}
