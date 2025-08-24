package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceIndexTemplate(t *testing.T) {
	// generate random template name
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceIndexTemplateDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
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
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:   testAccResourceIndexTemplateCreateWithIgnoreComponent(templateNameComponent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "index_patterns.*", fmt.Sprintf("%s-logscomponent-*", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "composed_of.*", fmt.Sprintf("%s-logscomponent@custom", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "ignore_missing_component_templates.*", fmt.Sprintf("%s-logscomponent@custom", templateNameComponent)),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				Config:   testAccResourceIndexTemplateUpdateWithIgnoreComponent(templateNameComponent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "index_patterns.*", fmt.Sprintf("%s-logscomponent-*", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "composed_of.*", fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent)),
					resource.TestCheckTypeSetElemAttr("elasticstack_elasticsearch_index_template.test_component", "ignore_missing_component_templates.*", fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent)),
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

  template {}
}
	`, name, name, name)
}

func testAccResourceIndexTemplateCreateWithIgnoreComponent(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
  name = "%s"
	index_patterns = ["%s-logscomponent-*"]

  composed_of = ["%s-logscomponent@custom"]
  ignore_missing_component_templates = ["%s-logscomponent@custom"]
}
	`, name, name, name, name)
}

func testAccResourceIndexTemplateUpdateWithIgnoreComponent(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index_template" "test_component" {
  name = "%s"
	index_patterns = ["%s-logscomponent-*"]

  composed_of = ["%s-logscomponent-updated@custom"]
  ignore_missing_component_templates = ["%s-logscomponent-updated@custom"]

  template {
  }
}
	`, name, name, name, name)
}

func checkResourceIndexTemplateDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_template" {
			continue
		}
		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		esClient, err := client.GetESClient()
		if err != nil {
			return err
		}
		req := esClient.Indices.GetIndexTemplate.WithName(compId.ResourceId)
		res, err := esClient.Indices.GetIndexTemplate(req)
		if err != nil {
			return err
		}

		if res.StatusCode != 404 {
			return fmt.Errorf("Index template (%s) still exists", compId.ResourceId)
		}
	}
	return nil
}
