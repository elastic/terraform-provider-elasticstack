package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceComponentTemplate(t *testing.T) {
	// generate a random username
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceComponentTemplateDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceComponentTemplateCreate(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.alias.0.name", "my_template_test"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "template.0.settings", `{"index":{"number_of_shards":"3"}}`),
				),
			},
		},
	})
}

func testAccResourceComponentTemplateCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_component_template" "test" {
  name = "%s"

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}`, name)
}

func checkResourceComponentTemplateDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_component_template" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Cluster.GetComponentTemplate.WithName(compId.ResourceId)
		res, err := client.GetESClient().Cluster.GetComponentTemplate(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Component template (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
