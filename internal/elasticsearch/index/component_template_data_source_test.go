package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceComponentTemplate(t *testing.T) {
	// generate a random username
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceComponentTemplateRead(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_component_template.test", "name", templateName),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_component_template.test", "template.0.alias.0.name", "my_template_test"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_component_template.test", "template.0.settings", `{"index":{"number_of_shards":"3"}}`),
				),
			},
		},
	})
}

func testAccDataSourceComponentTemplateRead(name string) string {
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
}

data "elasticstack_elasticsearch_component_template" "test" {
	name = elasticstack_elasticsearch_component_template.test.name
}
`, name)
}
