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
	"github.com/hashicorp/terraform-plugin-testing/config"
	tf_acctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceAgentConfiguration(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)
	expectedID := serviceName + ":production"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "id", expectedID),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "go"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.5"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "all"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "id", expectedID),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "production"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.capture_body", "off"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_settings"),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "id", expectedID),
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

func TestAccResourceAgentConfiguration_alternateEnvironment(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)
	expectedID := serviceName + ":staging"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "id", expectedID),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment", "staging"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name", "java"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "2"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.8"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.log_level", "debug"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				ResourceName:      "elasticstack_apm_agent_configuration.test_config",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedID := is[0].Attributes["id"]
					if importedID != expectedID {
						return fmt.Errorf("expected imported id [%s] to equal [%s]", importedID, expectedID)
					}

					return nil
				},
			},
		},
	})
}

func TestAccResourceAgentConfiguration_minimal(t *testing.T) {
	serviceName := tf_acctest.RandStringFromCharSet(10, tf_acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "id", serviceName),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_name", serviceName),
					resource.TestCheckNoResourceAttr("elasticstack_apm_agent_configuration.test_config", "service_environment"),
					resource.TestCheckNoResourceAttr("elasticstack_apm_agent_configuration.test_config", "agent_name"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.%", "1"),
					resource.TestCheckResourceAttr("elasticstack_apm_agent_configuration.test_config", "settings.transaction_sample_rate", "0.5"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory(""),
				ConfigVariables: config.Variables{
					"service_name": config.StringVariable(serviceName),
				},
				ResourceName:      "elasticstack_apm_agent_configuration.test_config",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					importedID := is[0].Attributes["id"]
					if importedID != serviceName {
						return fmt.Errorf("expected imported id [%s] to equal [%s]", importedID, serviceName)
					}

					return nil
				},
			},
		},
	})
}
