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
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var (
	minVersionOutput       = version.Must(version.NewVersion("8.6.0"))
	minVersionOutputSpaces = version.Must(version.NewVersion("9.1.0"))
)

//go:embed testdata/TestAccResourceOutputElasticsearchFromSDK/create/main.tf
var sdkCreateTestConfig string

func TestAccResourceOutputElasticsearchFromSDK(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				Config: sdkCreateTestConfig,
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-elasticsearch-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-elasticsearch-output", policyName)),
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

func TestAccResourceOutputElasticsearch(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-elasticsearch-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-elasticsearch-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ResourceName:      "elasticstack_fleet_output.test_output",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//go:embed testdata/TestAccResourceOutputLogstashFromSDK/create/main.tf
var logstashSDKCreateTestConfig string

func TestAccResourceOutputLogstashFromSDK(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.7",
					},
				},
				Config: logstashSDKCreateTestConfig,
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-logstash-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.certificate_authorities.0", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.certificate", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.key", "placeholder"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-logstash-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate_authorities.0", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.key", "placeholder"),
				),
			},
		},
	})
}

func TestAccResourceOutputLogstash(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-logstash-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate_authorities.0", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.key", "placeholder"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Logstash Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-logstash-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "logstash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "logstash:5044"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate_authorities.0", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.certificate", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.key", "placeholder"),
				),
			},
		},
	})
}

func TestAccResourceOutputKafka(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Kafka Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-kafka-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "kafka:9092"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.auth_type", "none"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.topic", "beats"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.partition", "hash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.compression", "gzip"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.compression_level", "6"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.connection_type", "plaintext"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.required_acks", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.0.key", "environment"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.0.value", "test"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Kafka Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-kafka-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "config_yaml", "\"ssl.verification_mode\": \"none\"\n"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "kafka:9092"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.auth_type", "none"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.topic", "logs"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.partition", "round_robin"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.compression", "snappy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.connection_type", "encryption"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.required_acks", "-1"),
				),
			},
		},
	})
}

func TestAccResourceOutputKafkaComplex(t *testing.T) {
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Complex Kafka Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.auth_type", "none"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.topic", "complex-topic"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.partition", "hash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.compression", "lz4"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.required_acks", "0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.broker_timeout", "10"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.timeout", "30"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.version", "2.6.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.0.key", "datacenter"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.0.value", "us-west-1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.1.key", "service"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.headers.1.value", "beats"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.hash.hash", "event.hash"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.hash.random", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.sasl.mechanism", "SCRAM-SHA-256"),
				),
			},
		},
	})
}

func TestAccResourceOutputRemoteElasticsearch(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	require.NoError(t, err)
	kibanaOapiClient, err := client.GetKibanaOapiClient()
	require.NoError(t, err)
	remote := true
	resp, err := kibanaOapiClient.API.PostFleetServiceTokensWithResponse(t.Context(), kbapi.PostFleetServiceTokensJSONRequestBody{
		Remote: &remote,
	})
	require.NoError(t, err)
	if resp == nil {
		t.Skip("skipping remote output acceptance test: no response when creating remote service token")
	}
	if resp.JSON200 == nil || strings.TrimSpace(resp.JSON200.Value) == "" {
		t.Skipf("skipping remote output acceptance test: unable to create remote service token (status=%d, body=%s)", resp.StatusCode(), string(resp.Body))
	}
	serviceToken := strings.Trim(strings.TrimSpace(resp.JSON200.Value), "\"")
	if !strings.Contains(serviceToken, ":") {
		t.Skipf("skipping remote output acceptance test: unexpected remote service token format (status=%d, body=%s)", resp.StatusCode(), string(resp.Body))
	}

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name":   config.StringVariable(policyName),
					"service_token": config.StringVariable(serviceToken),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Remote Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-remote-elasticsearch-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "remote_elasticsearch"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_output.test_output", "service_token"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "sync_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "sync_uninstalled_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "write_to_logs_streams", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name":   config.StringVariable(policyName),
					"service_token": config.StringVariable(serviceToken),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Updated Remote Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "id", fmt.Sprintf("%s-remote-elasticsearch-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "remote_elasticsearch"),
					resource.TestCheckResourceAttrSet("elasticstack_fleet_output.test_output", "service_token"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "sync_integrations", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "sync_uninstalled_integrations", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "write_to_logs_streams", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "hosts.0", "https://elasticsearch:9200"),
				),
			},
		},
	})
}

func TestAccResourceOutputRemoteElasticsearchValidation(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation-sync-integrations"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)sync_integrations.*remote_elasticsearch`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation-sync-uninstalled-integrations"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)sync_uninstalled_integrations.*remote_elasticsearch`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation-write-to-logs-streams"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)write_to_logs_streams.*remote_elasticsearch`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation-missing-service-token"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)service_token.*must be set.*remote_elasticsearch`),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation-service-token-on-elasticsearch"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				ExpectError: regexp.MustCompile(`(?s)service_token.*remote_elasticsearch`),
			},
		},
	})
}

func TestAccResourceFleetOutput_importFromSpace(t *testing.T) {
	versionutils.SkipIfUnsupported(t, minVersionOutputSpaces, versionutils.FlavorAny)

	policyName := sdkacctest.RandString(22)
	spaceName := sdkacctest.RandString(22)
	spaceID := fmt.Sprintf("fleet-output-test-%s", spaceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(fmt.Sprintf("Fleet Output Test Space %s", spaceName)),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "name", fmt.Sprintf("Elasticsearch Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_output.test_output", "space_ids.*", spaceID),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(fmt.Sprintf("Fleet Output Test Space %s", spaceName)),
				},
				ResourceName:            "elasticstack_fleet_output.test_output",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_output.test_output"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_output.test_output not found in state")
					}
					return fmt.Sprintf("%s/%s", spaceID, res.Primary.Attributes["output_id"]), nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "space_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_fleet_output.test_output", "space_ids.*", spaceID),
				),
			},
			// Scenario 2: plain ID import (no space prefix) - space_ids is not populated
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
					"space_id":    config.StringVariable(spaceID),
					"space_name":  config.StringVariable(fmt.Sprintf("Fleet Output Test Space %s", spaceName)),
				},
				ResourceName:            "elasticstack_fleet_output.test_output",
				ImportState:             true,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{"space_ids"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					res := s.RootModule().Resources["elasticstack_fleet_output.test_output"]
					if res == nil || res.Primary == nil {
						return "", fmt.Errorf("resource elasticstack_fleet_output.test_output not found in state")
					}
					return res.Primary.Attributes["output_id"], nil
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckNoResourceAttr("elasticstack_fleet_output.test_output", "space_ids.#"),
				),
			},
		},
	})
}

func TestAccResourceOutputElasticsearchWithFingerprint(t *testing.T) {
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ca_sha256", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ca_trusted_fingerprint", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891"),
				),
			},
		},
	})
}

func TestAccResourceOutputElasticsearchSSL(t *testing.T) {
	policyName := sdkacctest.RandString(22)
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputSSLVerificationMode, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.verification_mode", "none"),
				),
			},
		},
	})
}

func TestAccResourceOutputDefaultFlags(t *testing.T) {
	policyName := sdkacctest.RandString(22)
	versionutils.SkipIfUnsupported(t, minVersionOutput, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "true"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "true"),
				),
			},
			{
				// Kibana refuses to demote the cluster's only default output, so
				// promote the preconfigured `fleet-default-output` first. That
				// auto-demotes our test_output, after which Terraform's apply
				// happily aligns state with the now-non-default config.
				PreConfig:                promoteFleetDefaultOutput(t),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "elasticsearch"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_integrations", "false"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "default_monitoring", "false"),
				),
			},
		},
	})
}

// promoteFleetDefaultOutput returns a PreConfig hook that promotes the
// preconfigured `fleet-default-output` to be the cluster's default for
// integrations and monitoring. This is required when a managed test output is
// currently the default and we want to demote (or destroy) it: Kibana rejects
// demotions/deletions of an output that would leave the cluster without a
// default.
func promoteFleetDefaultOutput(t *testing.T) func() {
	return func() {
		t.Helper()
		client, err := clients.NewAcceptanceTestingKibanaScopedClient()
		require.NoError(t, err)
		kbClient, err := client.GetKibanaOapiClient()
		require.NoError(t, err)

		body := strings.NewReader(`{
			"name": "default",
			"type": "elasticsearch",
			"hosts": ["http://localhost:9200"],
			"is_default": true,
			"is_default_monitoring": true
		}`)
		resp, err := kbClient.API.PutFleetOutputsOutputidWithBodyWithResponse(
			t.Context(),
			"fleet-default-output",
			"application/json",
			body,
		)
		require.NoError(t, err)
		require.Equalf(t, 200, resp.StatusCode(),
			"failed to promote fleet-default-output to default: status=%d body=%s",
			resp.StatusCode(), string(resp.Body))
	}
}

func TestAccResourceOutputKafkaUserPass(t *testing.T) {
	policyName := sdkacctest.RandString(22)
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.auth_type", "user_pass"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.username", "testuser"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.sasl.mechanism", "PLAIN"),
				),
			},
		},
	})
}

func TestAccResourceOutputKafkaPartitions(t *testing.T) {
	policyName := sdkacctest.RandString(22)
	versionutils.SkipIfUnsupported(t, output.MinVersionOutputKafka, versionutils.FlavorAny)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.partition", "random"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.random.group_events", "1"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.partition", "round_robin"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "kafka.round_robin.group_events", "1"),
				),
			},
		},
	})
}

func checkResourceOutputDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
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
		spaceID := rs.Primary.Attributes["space_ids.0"]
		output, diags := fleet.GetOutput(context.Background(), fleetClient, rs.Primary.ID, spaceID)
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if output != nil {
			return fmt.Errorf("output id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
