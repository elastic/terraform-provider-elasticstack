package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var minVersionForFleet = version.Must(version.NewVersion("8.6.0"))

func TestProvider(t *testing.T) {
	if err := provider.New("dev").InternalValidate(); err != nil {
		t.Fatalf("Failed to validate provider: %s", err)
	}
}

func TestElasticsearchAPIKeyConnection(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(security.APIKeyMinVersion),
				Config:   testElasticsearchConnection(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_user.test", "username", "elastic"),
				),
			},
		},
	})
}

func TestFleetConfiguration(t *testing.T) {
	envConfig := config.NewFromEnv("acceptance-testing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionForFleet),
				Config:   testFleetConfiguration(envConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.elasticstack_fleet_enrollment_tokens.test", "tokens.#"),
				),
			},
		},
	})
}

func TestKibanaConfiguration(t *testing.T) {
	envConfig := config.NewFromEnv("acceptance-testing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testKibanaConfiguration(envConfig),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_space.acc_test", "name"),
				),
			},
		},
	})
}

func testKibanaConfiguration(cfg config.Client) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {
		endpoints = ["%s"]
		username  = "%s"
		password  = "%s"
	}
}

resource "elasticstack_kibana_space" "acc_test" {
	space_id          = "acc_test_space"
	name              = "Acceptance Test Space"
}`, cfg.Kibana.Address, cfg.Kibana.Username, cfg.Kibana.Password)
}

func testFleetConfiguration(cfg config.Client) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	fleet {
		endpoint = "%s"
		username = "%s"
		password = "%s"
	}
}

data "elasticstack_fleet_enrollment_tokens" "test" {}`, cfg.Fleet.URL, cfg.Fleet.Username, cfg.Fleet.Password)
}

func testElasticsearchConnection(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test_connection" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names = ["*"]
        privileges = ["all"]
        allow_restricted_indices = false
      }]
    }
  })

  expiration = "1d"
}


data "elasticstack_elasticsearch_security_user" "test" {
  username = "elastic"

  elasticsearch_connection {
    endpoints = ["%s"]
    api_key   = elasticstack_elasticsearch_security_api_key.test_connection.encoded
  }
}
`, apiKeyName, os.Getenv("ELASTICSEARCH_ENDPOINTS"))
}
