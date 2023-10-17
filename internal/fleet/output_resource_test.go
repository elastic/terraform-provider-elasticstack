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

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

func TestAccResourceOutputElasticsearch(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputCreateElasticsearch(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputUpdateElasticsearch(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
		},
	})
}

func TestAccResourceOutputLogstash(t *testing.T) {
	policyName := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputCreateLogstash(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputLogstashUpdate(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
				),
			},
		},
	})
}

func testAccResourceOutputCreateElasticsearch(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "elasticsearch"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}
`, fmt.Sprintf("Elasticsearch Output %s", id))
}

func testAccResourceOutputUpdateElasticsearch(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "elasticsearch"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "https://elasticsearch:9200"
  ]
}

`, fmt.Sprintf("Updated Elasticsearch Output %s", id))
}

func testAccResourceOutputCreateLogstash(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "logstash"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
}
`, fmt.Sprintf("Logstash Output %s", id))
}

func testAccResourceOutputLogstashUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "%s"
  type                 = "logstash"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
}

`, fmt.Sprintf("Updated Logstash Output %s", id))
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
