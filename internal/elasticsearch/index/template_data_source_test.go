package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIndexTemplateDataSource(t *testing.T) {
	// generate a random role name
	templateName := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := "test-template-" + sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccIndexTemplateDataSourceConfig(templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("tf-acc-%s-*", templateName)),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "priority", "100"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:   testAccIndexTemplateDataSourceWithIgnoreComponentConfig(templateNameComponent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_index_template.test", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("tf-acc-component-%s-*", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "composed_of.*", fmt.Sprintf("%s-logscomponent@custom", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_index_template.test", "ignore_missing_component_templates.*", fmt.Sprintf("%s-logscomponent@custom", templateNameComponent)),
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
}

data "elasticstack_elasticsearch_index_template" "test" {
	name = elasticstack_elasticsearch_index_template.test.name
}
	`, templateName, templateName)
}

func testAccIndexTemplateDataSourceWithIgnoreComponentConfig(templateName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
	name = "%s"
	index_patterns = ["tf-acc-component-%s-*"]
	composed_of = ["%s-logscomponent@custom"]
	ignore_missing_component_templates = ["%s-logscomponent@custom"]
}

data "elasticstack_elasticsearch_index_template" "test_component" {
	name = elasticstack_elasticsearch_index_template.test_component.name
}
	`, templateName, templateName, templateName, templateName)
}
