package security_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSecuritApiKey(t *testing.T) {
	// generate a random name
	apiKeyName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceSecurityApiKeyDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecuritApiKeyCreate(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "name", apiKeyName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", `{"role-a":{"cluster":["all"],"indices":[{"names":["index-a*"],"privileges":["read"],"allow_restricted_indices":false}]}}`),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "expiration"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "api_key"),
					resource.TestCheckResourceAttrSet("elasticstack_elasticsearch_security_api_key.test", "encoded"),
				),
			},
		},
	})
}

func testAccResourceSecuritApiKeyCreate(apiKeyName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "test" {
  name = "%s"

  role_descriptors = jsonencode({
    role-a = {
      cluster = ["all"]
      indices = [{
        names = ["index-a*"]
        privileges = ["read"]
        allow_restricted_indices = false
      }]
    }
  })

	expiration = "1d"
}
	`, apiKeyName)
}

func checkResourceSecurityApiKeyDestroy(s *terraform.State) error {
	client := acctest.ApiClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_security_api_key" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		apiKey, diags := client.GetElasticsearchApiKey(compId.ResourceId)
		if diags.HasError() {
			return fmt.Errorf("Unabled to get API key %v", diags)
		}

		if !apiKey.Invalidated {
			return fmt.Errorf("ApiKey (%s) has not been invalidated", compId.ResourceId)
		}
	}
	return nil
}
