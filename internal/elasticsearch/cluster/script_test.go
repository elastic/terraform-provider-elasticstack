package cluster_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceScript(t *testing.T) {
	scriptID := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkScriptDestroy,
		ProviderFactories: acctest.Providers,
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
  source    = "Math.log(_score * 4) + params['changed_modifier']"
  params    = jsonencode({
    changed_modifier = 2
  })
}
	`, id)
}

func checkScriptDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_script" {
			continue
		}

		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
		res, err := client.GetESClient().GetScript(compId.ResourceId)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("script (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
