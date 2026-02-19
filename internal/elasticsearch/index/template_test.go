package index_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceIndexTemplate(t *testing.T) {
	// generate random template name
	templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	templateNameComponent := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-logs-*", templateName),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "priority", "42"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "name", templateName),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test",
						"index_patterns.*",
						fmt.Sprintf("%s-logs-*", templateName),
					),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test", "template.0.alias.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "name", fmt.Sprintf("%s-stream", templateName)),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test2", "data_stream.0.hidden", "false"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create_ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("%s-logscomponent-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent@custom", templateNameComponent),
					),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index.MinSupportedIgnoreMissingComponentTemplateVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_ignore_component"),
				ConfigVariables: config.Variables{
					"template_name": config.StringVariable(templateNameComponent),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_index_template.test_component", "name", templateNameComponent),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"index_patterns.*",
						fmt.Sprintf("%s-logscomponent-*", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"composed_of.*",
						fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent),
					),
					resource.TestCheckTypeSetElemAttr(
						"elasticstack_elasticsearch_index_template.test_component",
						"ignore_missing_component_templates.*",
						fmt.Sprintf("%s-logscomponent-updated@custom", templateNameComponent),
					),
				),
			},
		},
	})
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
