package integration_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	minVersionIntegration       = version.Must(version.NewVersion("8.6.0"))
	minVersionIntegrationPolicy = version.Must(version.NewVersion("8.10.0"))
)

func TestAccResourceIntegrationFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:   testAccResourceIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:                   testAccResourceIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
		},
	})
}

func TestAccResourceIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:   testAccResourceIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				ResourceName:      "elasticstack_fleet_integration.test_integration",
				Config:            testAccResourceIntegration,
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("Resource Import Not Implemented"),
			},
		},
	})
}

func TestAccResourceIntegrationWithPolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationWithPolicy(policyName, "1.16.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.16.0"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				Config:   testAccResourceIntegrationWithPolicy(policyName, "1.17.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.17.0"),
				),
			},
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationPolicy),
				ResourceName:      "elasticstack_fleet_integration.test_integration",
				Config:            testAccResourceIntegrationWithPolicy(policyName, "1.17.0"),
				ImportState:       true,
				ImportStateVerify: true,
				ExpectError:       regexp.MustCompile("Resource Import Not Implemented"),
			},
		},
	})
}

func TestAccResourceIntegrationDeleted(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:   testAccResourceIntegrationDeleted,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "name", "sysmon_linux"),
					resource.TestCheckResourceAttr("elasticstack_fleet_integration.test_integration", "version", "1.7.0"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegration),
				Config:   testAccResourceIntegrationDeleted,
				// Force uninstall the integration
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					fleetClient, err := client.GetFleetClient()
					require.NoError(t, err)

					ctx := context.Background()
					diags := fleet.Uninstall(ctx, fleetClient, "sysmon_linux", "1.7.0", true)
					require.Empty(t, diags)
				},
				// Expect the plan to want to reinstall
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccResourceIntegration = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "1.16.0"
  force        = true
  skip_destroy = true
}
`

func testAccResourceIntegrationWithPolicy(policyName, version string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "tcp"
  version      = "%s"
  force        = true
  skip_destroy = true
}

// An agent policy to hold the integration policy.
resource "elasticstack_fleet_agent_policy" "sample" {
  name            = "%s"
  namespace       = "default"
  description     = "A sample agent policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

// The associated enrollment token.
data "elasticstack_fleet_enrollment_tokens" "sample" {
  policy_id = elasticstack_fleet_agent_policy.sample.policy_id
}

// The integration policy.
resource "elasticstack_fleet_integration_policy" "sample" {
  name                = "%s"
  namespace           = "default"
  description         = "A sample integration policy"
  agent_policy_id     = elasticstack_fleet_agent_policy.sample.policy_id
  integration_name    = elasticstack_fleet_integration.test_integration.name
  integration_version = elasticstack_fleet_integration.test_integration.version

  input {
    input_id = "tcp-tcp"
    streams_json = jsonencode({
      "tcp.generic" : {
        "enabled" : true,
        "vars" : {
          "listen_address" : "localhost",
          "listen_port" : 8080,
          "data_stream.dataset" : "tcp.generic",
          "tags" : [],
          "syslog_options" : "field: message\n#format: auto\n#timezone: Local\n",
          "ssl" : "#certificate: |\n#    -----BEGIN CERTIFICATE-----\n#    ...\n#    -----END CERTIFICATE-----\n#key: |\n#    -----BEGIN PRIVATE KEY-----\n#    ...\n#    -----END PRIVATE KEY-----\n",
          "custom" : ""
        }
      }
    })
  }
}
`, version, policyName, policyName)
}

const testAccResourceIntegrationDeleted = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_integration" "test_integration" {
  name         = "sysmon_linux"
  version      = "1.7.0"
  force        = true
  skip_destroy = false
}
`
