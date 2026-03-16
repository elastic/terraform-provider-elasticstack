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
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	tf_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccResourceAgentConfiguration_alternateEnvironment(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAgentConfigurationAlternateEnvironment(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "staging"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.log_level", "debug"),
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

func testAccResourceAgentConfigurationAlternateEnvironment(serviceName string) string {
	return fmt.Sprintf(`
	provider "elasticstack" {
		kibana {}
	}

	resource "elasticstack_apm_agent_configuration" "test_config" {
		service_name        = "%s"
		service_environment = "staging"
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
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.5"),
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

func TestAccResourceAgentConfiguration_updateServiceEnvironment(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceName, "production", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					testCheckAPMAgentConfigurationExists(serviceName, "production", true),
				),
			},
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceName, "staging", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "staging"),
					testCheckAPMAgentConfigurationExists(serviceName, "staging", true),
					testCheckAPMAgentConfigurationAbsent(serviceName, "production", true),
				),
			},
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceName, "", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckNoResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment"),
					testCheckAPMAgentConfigurationExists(serviceName, "", false),
					testCheckAPMAgentConfigurationAbsent(serviceName, "staging", true),
				),
			},
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceName, "production", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					testCheckAPMAgentConfigurationExists(serviceName, "production", true),
					testCheckAPMAgentConfigurationAbsent(serviceName, "", false),
				),
			},
		},
	})
}

func TestAccResourceAgentConfiguration_renameService(t *testing.T) {
	serviceNameOne := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)
	serviceNameTwo := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceNameOne, "production", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceNameOne),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					testCheckAPMAgentConfigurationExists(serviceNameOne, "production", true),
				),
			},
			{
				Config: testAccResourceAgentConfigurationLifecycle(serviceNameTwo, "production", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_apm_agent_configuration.test_config", "id"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceNameTwo),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					testCheckAPMAgentConfigurationExists(serviceNameTwo, "production", true),
					testCheckAPMAgentConfigurationAbsent(serviceNameOne, "production", true),
				),
			},
		},
	})
}

func testAccResourceAgentConfigurationLifecycle(serviceName, serviceEnvironment string, includeEnvironment bool) string {
	serviceEnvironmentConfig := ""
	if includeEnvironment {
		serviceEnvironmentConfig = fmt.Sprintf("\n\t\tservice_environment = %q", serviceEnvironment)
	}

	return fmt.Sprintf(`
	provider "elasticstack" {
		kibana {}
	}

	resource "elasticstack_apm_agent_configuration" "test_config" {
		service_name = "%s"%s
		agent_name   = "java"
		settings = {
			"transaction_sample_rate" = "0.8"
			"log_level"               = "debug"
		}
	}
	`, serviceName, serviceEnvironmentConfig)
}

func testCheckAPMAgentConfigurationExists(serviceName, serviceEnvironment string, hasEnvironment bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return testCheckAPMAgentConfiguration(serviceName, serviceEnvironment, hasEnvironment, true)
	}
}

func testCheckAPMAgentConfigurationAbsent(serviceName, serviceEnvironment string, hasEnvironment bool) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return testCheckAPMAgentConfiguration(serviceName, serviceEnvironment, hasEnvironment, false)
	}
}

func testCheckAPMAgentConfiguration(serviceName, serviceEnvironment string, hasEnvironment bool, shouldExist bool) error {
	configs, err := fetchAPMAgentConfigurations()
	if err != nil {
		return err
	}

	found := false
	for _, config := range configs {
		if config.Service.Name == nil || *config.Service.Name != serviceName {
			continue
		}

		if hasEnvironment {
			if config.Service.Environment == nil || *config.Service.Environment != serviceEnvironment {
				continue
			}
		} else if config.Service.Environment != nil && *config.Service.Environment != "" {
			continue
		}

		found = true
		break
	}

	if found == shouldExist {
		return nil
	}

	environmentDescription := "unset"
	if hasEnvironment {
		environmentDescription = serviceEnvironment
	}

	if shouldExist {
		return fmt.Errorf("expected APM agent configuration for service_name=%q service_environment=%q to exist", serviceName, environmentDescription)
	}

	return fmt.Errorf("expected APM agent configuration for service_name=%q service_environment=%q to be absent", serviceName, environmentDescription)
}

func fetchAPMAgentConfigurations() ([]kbapi.APMUIAgentConfigurationObject, error) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return nil, err
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return nil, err
	}

	resp, err := kibanaClient.API.GetAgentConfigurationsWithResponse(
		context.Background(),
		&kbapi.GetAgentConfigurationsParams{
			ElasticApiVersion: "2023-10-31",
		},
	)
	if err != nil {
		return nil, err
	}

	if resp.JSON200 == nil || resp.JSON200.Configurations == nil {
		return nil, fmt.Errorf("expected APM agent configurations response body to be populated")
	}

	return *resp.JSON200.Configurations, nil
}
