package output_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/output"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var minVersionOutput = version.Must(version.NewVersion("8.6.0"))

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
				Config:   testAccResourceOutputCreateElasticsearch(policyName),
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
				Config:                   testAccResourceOutputCreateElasticsearch(policyName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputCreateElasticsearch(policyName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputUpdateElasticsearch(policyName),
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
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:            testAccResourceOutputUpdateElasticsearch(policyName),
				ResourceName:      "elasticstack_fleet_output.test_output",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

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
				Config:   testAccResourceOutputCreateLogstash(policyName, true),
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
				Config:                   testAccResourceOutputCreateLogstash(policyName, false),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputCreateLogstash(policyName, false),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionOutput),
				Config:   testAccResourceOutputUpdateLogstash(policyName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
				Config:   testAccResourceOutputCreateKafka(policyName),
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
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
				Config:   testAccResourceOutputUpdateKafka(policyName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(output.MinVersionOutputKafka),
				Config:   testAccResourceOutputCreateKafkaComplex(policyName),
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

func testAccResourceOutputCreateElasticsearch(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Elasticsearch Output %s"
  output_id            = "%s-elasticsearch-output"
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
`, id, id)
}

func testAccResourceOutputUpdateElasticsearch(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Updated Elasticsearch Output %s"
  output_id            = "%s-elasticsearch-output"
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
`, id, id)
}

func testAccResourceOutputCreateLogstash(id string, forSDK bool) string {
	sslInfix := ""
	if !forSDK {
		sslInfix = "="
	}
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Logstash Output %s"
  type                 = "logstash"
  output_id            = "%s-logstash-output"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
  ssl %s {
	certificate_authorities = ["placeholder"]
	certificate             = "placeholder"
	key                     = "placeholder"
  }
}
`, id, id, sslInfix)
}

func testAccResourceOutputUpdateLogstash(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Updated Logstash Output %s"
  output_id            = "%s-logstash-output"
  type                 = "logstash"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "logstash:5044"
  ]
  ssl = {
	certificate_authorities = ["placeholder"]
	certificate             = "placeholder"
	key                     = "placeholder"
  }
}
`, id, id)
}

func testAccResourceOutputCreateKafka(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Kafka Output %s"
  output_id            = "%s-kafka-output"
  type                 = "kafka"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka:9092"
  ]
  
  # Kafka-specific configuration
  kafka = {
	auth_type         = "none"
	topic             = "beats"
	partition         = "hash"
	compression       = "gzip"
	compression_level = 6
	connection_type   = "plaintext"
	required_acks     = 1
  
	headers = [{
		key   = "environment"
		value = "test"
	}]
  }
}
`, id, id)
}

func testAccResourceOutputUpdateKafka(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Updated Kafka Output %s"
  output_id            = "%s-kafka-output"
  type                 = "kafka"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka:9092"
  ]
  
  # Updated Kafka-specific configuration
  kafka = {
	auth_type         = "none"
	topic             = "logs"
	partition         = "round_robin"
	compression       = "snappy"
	connection_type   = "encryption"
	required_acks     = -1
  }
}
`, id, id)
}

func testAccResourceOutputCreateKafkaComplex(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_fleet_output" "test_output" {
  name                 = "Complex Kafka Output %s"
  output_id            = "%s-kafka-complex-output"
  type                 = "kafka"
  config_yaml = yamlencode({
    "ssl.verification_mode" : "none"
  })
  default_integrations = false
  default_monitoring   = false
  hosts = [
    "kafka1:9092",
    "kafka2:9092",
    "kafka3:9092"
  ]
  
  # Complex Kafka configuration showcasing all options
  kafka = {
    auth_type         = "none"
    topic             = "complex-topic"
    partition         = "hash"
    compression       = "lz4"
    connection_type   = "encryption"
    required_acks     = 0
    broker_timeout    = 10
    timeout           = 30
    version           = "2.6.0"
  
    headers = [
      {
        key   = "datacenter"
        value = "us-west-1"
      },
      {
        key   = "service"
        value = "beats"
      }
    ]
  
    hash = {
      hash   = "event.hash"
      random = false
    }
  
    sasl = {
      mechanism = "SCRAM-SHA-256"
    }
  }
}
`, id, id)
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
