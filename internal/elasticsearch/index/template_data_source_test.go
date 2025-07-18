package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	// generate a random role name
	templateName := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceConfig(templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("tf-acc-%s-*", templateName)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", fmt.Sprintf("%s-logs@custom", templateName)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", fmt.Sprintf("%s-logs@custom", templateName)),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "priority", "100"),
				),
			},
		},
	})
}

func testAccIndexTemplateDataSourceConfig(templateName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
	name = "%s"

	priority       = 100
	index_patterns = ["tf-acc-%s-*"]

	composed_of = ["%s-logs@custom"]
	ignore_missing_component_templates = ["%s-logs@custom"]
}

data "elasticstack_elasticsearch_index_template" "test" {
	name = elasticstack_elasticsearch_index_template.test.name
}
	`, templateName, templateName, templateName, templateName)
}
