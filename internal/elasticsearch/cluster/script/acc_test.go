package script_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccResourceScript(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkScriptDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccScriptCreate(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
			{
				Config: testAccScriptUpdate(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
				),
			},
			{
				// Ensure the provider doesn't panic if the script has been deleted outside of the Terraform flow
				PreConfig: func() {
					client, err := clients.NewAcceptanceTestingClient()
					require.NoError(t, err)

					esClient, err := client.GetESClient()
					require.NoError(t, err)

					_, err = esClient.DeleteScript(scriptID)
					require.NoError(t, err)
				},
				Config: testAccScriptUpdate(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 4) + params['changed_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "params", `{"changed_modifier":2}`),
				),
			},
		},
	})
}

func TestAccResourceScriptSearchTemplate(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkScriptDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchTemplateCreate(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "lang", "mustache"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "source", `{"from":"{{from}}","query":{"match":{"message":"{{query_string}}"}},"size":"{{size}}"}`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.search_template_test", "params", `{"query_string":"My query string"}`),
				),
			},
		},
	})
}

func TestAccResourceScriptFromSDK(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the script with the last provider version where the script resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.17",
					},
				},
				Config: testAccScriptCreateFromSDK(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccScriptCreateFromSDK(scriptID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "script_id", scriptID),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "lang", "painless"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "source", "Math.log(_score * 2) + params['my_modifier']"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_script.test", "context", "score"),
				),
			},
		},
	})
}

func testAccScriptCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "test" {
  script_id = "%s"
  lang      = "painless"
  source    = "Math.log(_score * 2) + params['my_modifier']"
  context   = "score"
}
	`, id)
}

func testAccScriptUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "test" {
  script_id = "%s"
  lang      = "painless"
  source    = "Math.log(_score * 4) + params['changed_modifier']"
  params    = jsonencode({
    changed_modifier = 2
  })
}
	`, id)
}

func testAccSearchTemplateCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "search_template_test" {
  script_id = "%s"
  lang      = "mustache"
  source    = jsonencode({
    query = {
      match = {
        message = "{{query_string}}"
      }
    }
    from = "{{from}}"
    size = "{{size}}"
  })
  params = jsonencode({
    query_string = "My query string"
  })
}
	`, id)
}

func testAccScriptCreateFromSDK(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_script" "test" {
  script_id = "%s"
  lang      = "painless"
  source    = "Math.log(_score * 2) + params['my_modifier']"
  context   = "score"
}
	`, id)
}

func checkScriptDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_script" {
			continue
		}

		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		res, err := esClient.GetScript(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("script (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}