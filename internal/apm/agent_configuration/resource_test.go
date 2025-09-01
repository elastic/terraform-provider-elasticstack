package agent_configuration_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	tf_acctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceAgentConfiguration(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAgentConfigurationCreate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "go"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "all"),
				),
			},
			{
				Config: testAccResourceAgentConfigurationUpdate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "off"),
				),
			},
		},
	})
}

func testAccResourceAgentConfigurationCreate(serviceName string) string {
	return fmt.Sprintf(`
	provider "elasticstack" {
		kibana {}
	}

	resource "elasticstack_apm_agent_configuration" "test_config" {
		service_name        = "%s"
		service_environment = "production"
		agent_name          = "go"
		settings = {
			"transaction_sample_rate" = "0.5"
			"capture_body"            = "all"
		}
	}
	`, serviceName)
}

func testAccResourceAgentConfigurationUpdate(serviceName string) string {
	return fmt.Sprintf(`
	provider "elasticstack" {
		kibana {}
	}

	resource "elasticstack_apm_agent_configuration" "test_config" {
		service_name        = "%s"
		service_environment = "production"
		agent_name          = "java"
		settings = {
			"transaction_sample_rate" = "0.8"
			"capture_body"            = "off"
		}
	}
	`, serviceName)
}
