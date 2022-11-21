package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProvider(t *testing.T) {
	if err := acctest.Provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestElasticsearchAPIKeyConnection(t *testing.T) {
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.Providers,
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
