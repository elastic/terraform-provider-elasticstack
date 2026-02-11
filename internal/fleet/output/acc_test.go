package output_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

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
)

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

//go:embed testdata/TestAccResourceOutputElasticsearchFromSDK/create/output.tf
var sdkCreateTestConfig string

func TestAccResourceOutputElasticsearchFromSDK(t *testing.T) {
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   sdkCreateTestConfig,
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
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

func TestAccDataSourceFleetOutput(t *testing.T) {
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable(policyName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"id",
						fmt.Sprintf("default/%s-elasticsearch-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"output_id",
						fmt.Sprintf("%s-elasticsearch-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"name",
						fmt.Sprintf("Elasticsearch Output %s", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"type",
						"elasticsearch",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"hosts.0",
						"https://elasticsearch:9200",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"ca_sha256",
						"sha256fingerprint",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"ca_trusted_fingerprint",
						"trustedfingerprint",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"config_yaml",
						"\"ssl.verification_mode\": \"none\"\n",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"default_integrations",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.elasticsearch",
						"default_monitoring",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"id",
						fmt.Sprintf("default/%s-logstash-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"output_id",
						fmt.Sprintf("%s-logstash-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"name",
						fmt.Sprintf("Logstash Output %s", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"type",
						"logstash",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"hosts.0",
						"logstash:5044",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"ssl.certificate_authorities.0",
						"placeholder",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"ssl.certificate",
						"placeholder",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.logstash",
						"config_yaml",
						"\"ssl.verification_mode\": \"none\"\n",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"id",
						fmt.Sprintf("default/%s-kafka-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"output_id",
						fmt.Sprintf("%s-kafka-output", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"name",
						fmt.Sprintf("Kafka Output %s", policyName),
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"type",
						"kafka",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"hosts.0",
						"kafka:9092",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.auth_type",
						"none",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.topic",
						"beats",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.partition",
						"hash",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.compression",
						"gzip",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.compression_level",
						"6",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.connection_type",
						"plaintext",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.required_acks",
						"1",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.headers.0.key",
						"environment",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"kafka.headers.0.value",
						"test",
					),
					resource.TestCheckResourceAttr(
						"data.elasticstack_fleet_output.kafka",
						"config_yaml",
						"\"ssl.verification_mode\": \"none\"\n",
					),
				),
			},
		},
	})
}

//go:embed testdata/TestAccResourceOutputLogstashFromSDK/create/output.tf
var logstashSDKCreateTestConfig string

func TestAccResourceOutputLogstashFromSDK(t *testing.T) {
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   logstashSDKCreateTestConfig,
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
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
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
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
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceOutputDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
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
		output, diags := fleet.GetOutput(context.Background(), fleetClient, rs.Primary.ID, "")
		if diags.HasError() {
			return diagutil.FwDiagsAsError(diags)
		}
		if output != nil {
			return fmt.Errorf("output id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}
