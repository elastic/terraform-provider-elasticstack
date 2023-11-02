package fleet_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
)

var minVersionPackagePolicy = version.Must(version.NewVersion("8.10.0"))

func TestAccResourcePackagePolicy(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourcePackagePolicyDestroy,
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionPackagePolicy),
				Config:   testAccResourcePackagePolicyCreate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "description", "PackagePolicyTest Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "package_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "package_version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.input_id", "tcp-tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.streams_json", "{\"tcp.generic\":{\"enabled\":true,\"vars\":{\"custom\":\"\",\"data_stream.dataset\":\"tcp.generic\",\"listen_address\":\"localhost\",\"listen_port\":8080,\"ssl\":\"#certificate: |\\n#    -----BEGIN CERTIFICATE-----\\n#    ...\\n#    -----END CERTIFICATE-----\\n#key: |\\n#    -----BEGIN PRIVATE KEY-----\\n#    ...\\n#    -----END PRIVATE KEY-----\\n\",\"syslog_options\":\"field: message\\n#format: auto\\n#timezone: Local\\n\",\"tags\":[]}}}"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionPackagePolicy),
				Config:   testAccResourcePackagePolicyUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "name", policyName),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "description", "Updated Package Policy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "package_name", "tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "package_version", "1.16.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.input_id", "tcp-tcp"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.enabled", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_package_policy.test_policy", "input.0.streams_json", "{\"tcp.generic\":{\"enabled\":true,\"vars\":{\"custom\":\"\",\"data_stream.dataset\":\"tcp.generic\",\"listen_address\":\"localhost\",\"listen_port\":8085,\"ssl\":\"#certificate: |\\n#    -----BEGIN CERTIFICATE-----\\n#    ...\\n#    -----END CERTIFICATE-----\\n#key: |\\n#    -----BEGIN PRIVATE KEY-----\\n#    ...\\n#    -----END PRIVATE KEY-----\\n\",\"syslog_options\":\"field: message\\n#format: auto\\n#timezone: Local\\n\",\"tags\":[]}}}"),
				),
			},
		},
	})
}

func checkResourcePackagePolicyDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	fleetClient, err := client.GetFleetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "elasticstack_fleet_package_policy":
			packagePolicy, diag := fleet.ReadPackagePolicy(context.Background(), fleetClient, rs.Primary.ID)
			if diag.HasError() {
				return fmt.Errorf(diag[0].Summary)
			}
			if packagePolicy != nil {
				return fmt.Errorf("package policy id=%v still exists, but it should have been removed", rs.Primary.ID)
			}
		case "elasticstack_fleet_agent_policy":
			agentPolicy, diag := fleet.ReadAgentPolicy(context.Background(), fleetClient, rs.Primary.ID)
			if diag.HasError() {
				return fmt.Errorf(diag[0].Summary)
			}
			if agentPolicy != nil {
				return fmt.Errorf("agent policy id=%v still exists, but it should have been removed", rs.Primary.ID)
			}
		default:
			continue
		}

	}
	return nil
}

func testAccResourcePackagePolicyCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_package" "test_policy" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "%s Agent Policy"
  namespace       = "default"
  description     = "PackagePolicyTest Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

resource "elasticstack_fleet_package_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "PackagePolicyTest Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  package_name    = elasticstack_fleet_package.test_policy.name
  package_version = elasticstack_fleet_package.test_policy.version

  input {
    input_id = "tcp-tcp"
	streams_json = jsonencode({
	  "tcp.generic": {
	    "enabled": true,
	    "vars": {
	  	  "listen_address": "localhost",
  		  "listen_port": 8080, 
		  "data_stream.dataset": "tcp.generic",
		  "tags": [],
		  "syslog_options": "field: message\n#format: auto\n#timezone: Local\n",
		  "ssl": "#certificate: |\n#    -----BEGIN CERTIFICATE-----\n#    ...\n#    -----END CERTIFICATE-----\n#key: |\n#    -----BEGIN PRIVATE KEY-----\n#    ...\n#    -----END PRIVATE KEY-----\n",
		  "custom": ""
        }
	  }
	})
  }
}
`, id, id)
}

func testAccResourcePackagePolicyUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_package" "test_policy" {
  name    = "tcp"
  version = "1.16.0"
  force   = true
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name            = "%s Agent Policy"
  namespace       = "default"
  description     = "PackagePolicyTest Agent Policy"
  monitor_logs    = true
  monitor_metrics = true
  skip_destroy    = false
}

data "elasticstack_fleet_enrollment_tokens" "test_policy" {
  policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
}

resource "elasticstack_fleet_package_policy" "test_policy" {
  name            = "%s"
  namespace       = "default"
  description     = "Updated Package Policy"
  agent_policy_id = elasticstack_fleet_agent_policy.test_policy.policy_id
  package_name    = elasticstack_fleet_package.test_policy.name
  package_version = elasticstack_fleet_package.test_policy.version

  input {
    input_id = "tcp-tcp"
	streams_json = jsonencode({
	  "tcp.generic": {
	    "enabled": true,
	    "vars": {
	  	  "listen_address": "localhost",
  		  "listen_port": 8085, 
		  "data_stream.dataset": "tcp.generic",
		  "tags": [],
		  "syslog_options": "field: message\n#format: auto\n#timezone: Local\n",
		  "ssl": "#certificate: |\n#    -----BEGIN CERTIFICATE-----\n#    ...\n#    -----END CERTIFICATE-----\n#key: |\n#    -----BEGIN PRIVATE KEY-----\n#    ...\n#    -----END PRIVATE KEY-----\n",
		  "custom": ""
        }
	  }
	})
  }
}
`, id, id)
}
