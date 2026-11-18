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

package output_test

import (
	"fmt"
	"maps"
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceOutputElasticsearchPreset(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputPreset, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)
	vars := func(extra config.Variables) config.Variables {
		v := config.Variables{"policy_name": config.StringVariable(policyName)}
		maps.Copy(v, extra)
		return v
	}
	const resourceName = "elasticstack_fleet_output.test_output"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				// Fleet defaults the preset when it is omitted.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(nil),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("%s-elasticsearch-preset-output", policyName)),
					resource.TestCheckResourceAttr(resourceName, "preset", "balanced"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"preset": config.StringVariable("scale")}),
				Check:                    resource.TestCheckResourceAttr(resourceName, "preset", "scale"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"preset": config.StringVariable("scale")}),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
			},
			{
				// An unrelated change keeps the stored preset when it is no
				// longer configured and config_yaml is unchanged.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"host": config.StringVariable("https://elasticsearch-updated:9200")}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "hosts.0", "https://elasticsearch-updated:9200"),
					resource.TestCheckResourceAttr(resourceName, "preset", "scale"),
				),
			},
			{
				// Changing config_yaml without a configured preset makes Fleet
				// derive the custom preset from the performance keys.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"config_yaml": config.StringVariable("bulk_max_size: 100\n")}),
				Check:                    resource.TestCheckResourceAttr(resourceName, "preset", "custom"),
			},
			{
				// Fleet rejects balanced when config_yaml contains performance keys.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables: vars(config.Variables{
					"preset":      config.StringVariable("balanced"),
					"config_yaml": config.StringVariable("bulk_max_size: 100\n"),
				}),
				ExpectError: regexp.MustCompile(`(?s)preset.*balanced.*config_yaml`),
			},
			{
				// Only balanced conflicts with performance keys.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables: vars(config.Variables{
					"preset":      config.StringVariable("latency"),
					"config_yaml": config.StringVariable("bulk_max_size: 100\n"),
				}),
				Check: resource.TestCheckResourceAttr(resourceName, "preset", "latency"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables: vars(config.Variables{
					"preset":      config.StringVariable("throughput"),
					"config_yaml": config.StringVariable("\"ssl.verification_mode\": \"none\"\n"),
				}),
				Check: resource.TestCheckResourceAttr(resourceName, "preset", "throughput"),
			},
			{
				// custom is accepted without performance keys in config_yaml.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables: vars(config.Variables{
					"preset":      config.StringVariable("custom"),
					"config_yaml": config.StringVariable("\"ssl.verification_mode\": \"none\"\n"),
				}),
				Check: resource.TestCheckResourceAttr(resourceName, "preset", "custom"),
			},
		},
	})
}

func TestAccResourceOutputRemoteElasticsearchPreset(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputPreset, versionutils.FlavorAny)

	// Fleet stores the service token without validating it against the
	// remote cluster, so a placeholder is sufficient to exercise presets.
	serviceToken := "placeholder-remote-service-token"
	policyName := sdkacctest.RandString(22)
	vars := func(extra config.Variables) config.Variables {
		v := config.Variables{
			"policy_name":   config.StringVariable(policyName),
			"service_token": config.StringVariable(serviceToken),
		}
		maps.Copy(v, extra)
		return v
	}
	const resourceName = "elasticstack_fleet_output.test_output"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				// Creating without a preset succeeds whether or not this Fleet
				// version defaults it for remote_elasticsearch outputs.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(nil),
				Check:                    resource.TestCheckResourceAttr(resourceName, "type", "remote_elasticsearch"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"preset": config.StringVariable("latency")}),
				Check:                    resource.TestCheckResourceAttr(resourceName, "preset", "latency"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"preset": config.StringVariable("scale")}),
				Check:                    resource.TestCheckResourceAttr(resourceName, "preset", "scale"),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(config.Variables{"preset": config.StringVariable("scale")}),
				ResourceName:             resourceName,
				ImportState:              true,
				ImportStateVerify:        true,
				ImportStateVerifyIgnore:  []string{"service_token"},
			},
			{
				// Removing the preset from the configuration keeps the stored
				// value, which the provider sends back to Fleet.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("output"),
				ConfigVariables:          vars(nil),
				Check:                    resource.TestCheckResourceAttr(resourceName, "preset", "scale"),
			},
			{
				// Fleet re-derives the preset when the output type changes.
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("elasticsearch"),
				ConfigVariables:          config.Variables{"policy_name": config.StringVariable(policyName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "elasticsearch"),
					resource.TestCheckResourceAttr(resourceName, "preset", "balanced"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("logstash"),
				ConfigVariables:          config.Variables{"policy_name": config.StringVariable(policyName)},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "logstash"),
					resource.TestCheckNoResourceAttr(resourceName, "preset"),
				),
			},
		},
	})
}

func TestAccResourceOutputPresetValidation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputPreset, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("logstash"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(sdkacctest.RandString(22)),
				},
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`(?s)preset.*elasticsearch`),
			},
		},
	})
}
