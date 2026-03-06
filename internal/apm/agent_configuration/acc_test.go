// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package agentconfiguration_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	tf_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "go"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "all"),
				),
			},
			{
				Config: testAccResourceAgentConfigurationUpdate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "off"),
				),
			},
			{
				Config: testAccResourceAgentConfigurationUpdateSettings(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.log_level", "debug"),
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

func testAccResourceAgentConfigurationUpdateSettings(serviceName string) string {
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
			"log_level"               = "debug"
		}
	}
	`, serviceName)
}

func TestAccResourceAgentConfiguration_minimal(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAgentConfigurationMinimal(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckNoResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment"),
					resource.TestCheckNoResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "1"),
				),
			},
			{
				ResourceName:      "elasticstack_apm_agent_configuration.test_config",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceAgentConfigurationMinimal(serviceName string) string {
	return fmt.Sprintf(`
	provider "elasticstack" {
		kibana {}
	}

	resource "elasticstack_apm_agent_configuration" "test_config" {
		service_name = "%s"
		settings = {
			"transaction_sample_rate" = "0.5"
		}
	}
	`, serviceName)
}
