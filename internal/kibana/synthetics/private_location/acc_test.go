package private_location_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
  	elasticsearch {}
	kibana {}
}
`
)

var (
	minKibanaVersion = version.Must(version.NewVersion("8.12.0"))
)

func TestPrivateLocationResource(t *testing.T) {
	resourceId := "elasticstack_kibana_synthetics_private_location.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   testConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceId, "label", "test label"),
					resource.TestCheckResourceAttr(resourceId, "space_id", "testacc"),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "b"),
					resource.TestCheckResourceAttr(resourceId, "geo.lat", "42.42"),
					resource.TestCheckResourceAttr(resourceId, "geo.lon", "-42.42"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
				Config:            testConfig(),
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaVersion),
				Config:   updateConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceId, "label", "test label 2"),
					resource.TestCheckResourceAttr(resourceId, "space_id", "default"),
					resource.TestCheckResourceAttrSet(resourceId, "agent_policy_id"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceId, "tags.2", "e"),
					resource.TestCheckResourceAttr(resourceId, "geo.lat", "-33.21"),
					resource.TestCheckResourceAttr(resourceId, "geo.lon", "-33.21"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func agentPolicyResource(namespace, name string) string {
	return fmt.Sprintf(`
resource "elasticstack_fleet_agent_policy" "%s" {
	name            = "Private Location Agent Policy - %s"
	namespace       = "%s"
	description     = "TestPrivateLocationResource Agent Policy"
	monitor_logs    = true
	monitor_metrics = true
	skip_destroy    = false
}
`, name, name, namespace)
}

func testConfig() string {
	namespace := "testacc"
	return providerConfig + agentPolicyResource(namespace, "test_policy") +
		fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "%s" {
	label = "test label"
	space_id = "%s"
	agent_policy_id = %s
	tags = ["a", "b"]
	geo = {
		lat = 42.42
		lon = -42.42
	}
}
`, "test", namespace, "elasticstack_fleet_agent_policy.test_policy.policy_id")
}

func updateConfig() string {
	namespace := "default"
	return providerConfig + agentPolicyResource(namespace, "test_policy_default") +
		fmt.Sprintf(`
resource "elasticstack_kibana_synthetics_private_location" "%s" {
	label = "test label 2"
	space_id = "%s"
	agent_policy_id = %s
	tags = ["c", "d", "e"]
	geo = {
		lat = -33.21
		lon = -33.21
	}
}
`, "test", namespace, "elasticstack_fleet_agent_policy.test_policy_default.policy_id")
}
