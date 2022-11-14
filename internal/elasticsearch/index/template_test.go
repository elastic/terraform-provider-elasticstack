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

func TestAccResourceIndexTemplate(t *testing.T) {
	// generate random template name
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		CheckDestroy:      checkResourceIndexTemplateDestroy,
		ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceIndexTemplateCreate(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-logs-*", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "priority", "42"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "true"),
				),
			},
			{
				Config: testAccResourceIndexTemplateUpdate(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test", "index_patterns.*", fmt.Sprintf("%s-logs-*", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "false"),
				),
			},
		},
	})
}

func testAccResourceIndexTemplateCreate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = "%s"

  priority       = 42
  index_patterns = ["%s-logs-*"]

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "test2" {
  name = "%s-stream"

  index_patterns = ["index-pattern-streams*"]
  data_stream {
    hidden = true
  }
}
	`, name, name, name)
}

func testAccResourceIndexTemplateUpdate(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test" {
  name = "%s"

  index_patterns = ["%s-logs-*"]

  template {
    alias {
      name = "my_template_test"
    }
    alias {
      name = "alias2"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "test2" {
  name = "%s-stream"

  index_patterns = ["index-pattern-streams*"]
  data_stream {
    hidden = false
  }
}
	`, name, name, name)
}

func checkResourceIndexTemplateDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*clients.ApiClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_template" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		req := client.GetESClient().Indices.GetIndexTemplate.WithName(compId.ResourceId)
		res, err := client.GetESClient().Indices.GetIndexTemplate(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index template (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
