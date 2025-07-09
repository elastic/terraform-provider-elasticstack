package output_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	minVersionOutput      = version.Must(version.NewVersion("8.6.0"))
	minVersionKafkaOutput = version.Must(version.NewVersion("8.10.0"))
)

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
				Config:   testAccResourceOutputCreateLogstash(policyName),
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
				Config:                   testAccResourceOutputCreateLogstash(policyName),
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
				Config:   testAccResourceOutputCreateLogstash(policyName),
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
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.certificate_authorities.0", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.certificate", "placeholder"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_output", "ssl.0.key", "placeholder"),
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

func testAccResourceOutputCreateLogstash(id string) string {
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
  ssl {
	certificate_authorities = ["placeholder"]
	certificate             = "placeholder"
	key                     = "placeholder"
  }
}
`, id, id)
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
  ssl {
	certificate_authorities = ["placeholder"]
	certificate             = "placeholder"
	key                     = "placeholder"
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
		output, diags := fleet.GetOutput(context.Background(), fleetClient, rs.Primary.ID)
		if diags.HasError() {
			return utils.FwDiagsAsError(diags)
		}
		if output != nil {
			return fmt.Errorf("output id=%v still exists, but it should have been removed", rs.Primary.ID)
		}
	}
	return nil
}

func TestAccResourceOutputKafka(t *testing.T) {
	policyName := sdkacctest.RandString(22)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceOutputDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionKafkaOutput),
				Config:   testAccResourceOutputCreateKafka(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "name", fmt.Sprintf("Kafka Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "output_id", fmt.Sprintf("%s-kafka-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "hosts.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "hosts.0", "kafka-broker-1:9092"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.topic", "fleet-events"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.client_id", "tf-provider-test"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.version", "2.0.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.compression", "gzip"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.mechanism", "PLAIN"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.username", "user"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.password", "password"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionKafkaOutput),
				Config:   testAccResourceOutputUpdateKafka(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "name", fmt.Sprintf("Updated Kafka Output %s", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "output_id", fmt.Sprintf("%s-kafka-output", policyName)),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "type", "kafka"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "hosts.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "hosts.0", "kafka-broker-2:9092"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "hosts.1", "kafka-broker-3:9092"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.topic", "fleet-events-updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.client_id", "tf-provider-test-updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.version", "2.1.0"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.compression", "snappy"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.mechanism", "scram-sha-256"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.username", "user-updated"),
					resource.TestCheckResourceAttr("elasticstack_fleet_output.test_kafka_output", "kafka.0.sasl.0.password", "password-updated"),
				),
			},
		},
	})
}

func testAccResourceOutputCreateKafka(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_output" "test_kafka_output" {
  name      = "Kafka Output %s"
  output_id = "%s-kafka-output"
  type      = "kafka"
  hosts     = ["kafka-broker-1:9092"]

  kafka {
    topic       = "fleet-events"
    client_id   = "tf-provider-test"
    version     = "2.0.0"
    compression = "gzip"
    sasl {
      mechanism = "PLAIN"
      username  = "user"
      password  = "password"
    }
  }
}
`, id, id)
}

func testAccResourceOutputUpdateKafka(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_fleet_output" "test_kafka_output" {
  name      = "Updated Kafka Output %s"
  output_id = "%s-kafka-output"
  type      = "kafka"
  hosts     = ["kafka-broker-2:9092", "kafka-broker-3:9092"]

  kafka {
    topic       = "fleet-events-updated"
    client_id   = "tf-provider-test-updated"
    version     = "2.1.0"
    compression = "snappy"
    sasl {
      mechanism = "scram-sha-256"
      username  = "user-updated"
      password  = "password-updated"
    }
  }
}
`, id, id)
}
