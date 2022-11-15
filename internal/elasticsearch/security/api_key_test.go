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
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_api_key.test", "role_descriptors", `{"role-a":{"cluster":["all"],"index":[{"names":["index-a*"],"privileges":["read"]}]}}`),
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
      cluster = ["all"],
      index = [{
        names = ["index-a*"],
        privileges = ["read"]
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

		req := client.GetESClient().Security.GetAPIKey.WithID(compId.ResourceId)
		res, err := client.GetESClient().Security.GetAPIKey(req)
		if err != nil {
			return err
		}

		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("ApiKey (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
