package index_template_ilm_attachment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/index_template_ilm_attachment"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccResourceIndexTemplateIlmAttachment_fleet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index_template_ilm_attachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable("test-fleet-policy-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"index_template", "logs-tcp.generic"),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"lifecycle_name", "test-fleet-policy-1"),
					checkComponentTemplateExists("logs-tcp.generic@custom"),
					checkComponentTemplateHasILM("logs-tcp.generic@custom", "test-fleet-policy-1"),
				),
			},
			// Update
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index_template_ilm_attachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable("test-fleet-policy-2"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"lifecycle_name", "test-fleet-policy-2"),
					checkComponentTemplateHasILM("logs-tcp.generic@custom", "test-fleet-policy-2"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(index_template_ilm_attachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable("test-fleet-policy-2"),
				},
				ResourceName:      "elasticstack_elasticsearch_index_template_ilm_attachment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_template_ilm_attachment" {
			continue
		}

		compId, sdkDiags := clients.CompositeIdFromStr(rs.Primary.ID)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to parse resource ID: %v", sdkDiags)
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(context.Background(), client, compId.ResourceId)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to get component template: %v", sdkDiags)
		}

		// If the template still exists, check if ILM setting is removed
		if tpl != nil {
			if tpl.ComponentTemplate.Template != nil && tpl.ComponentTemplate.Template.Settings != nil {
				if _, hasILM := tpl.ComponentTemplate.Template.Settings["index.lifecycle.name"]; hasILM {
					return fmt.Errorf("ILM setting still exists in component template %s", compId.ResourceId)
				}
			}
		}
	}

	return nil
}

func checkComponentTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(context.Background(), client, name)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to get component template: %v", sdkDiags)
		}

		if tpl == nil {
			return fmt.Errorf("component template %s does not exist", name)
		}

		return nil
	}
}

func checkComponentTemplateHasILM(name string, expectedPolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(context.Background(), client, name)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to get component template: %v", sdkDiags)
		}

		if tpl == nil {
			return fmt.Errorf("component template %s does not exist", name)
		}

		if tpl.ComponentTemplate.Template == nil {
			return fmt.Errorf("component template %s has no template section", name)
		}

		if tpl.ComponentTemplate.Template.Settings == nil {
			return fmt.Errorf("component template %s has no settings", name)
		}

		// Elasticsearch returns settings in nested structure
		var actualPolicy string
		if indexSettings, ok := tpl.ComponentTemplate.Template.Settings["index"].(map[string]interface{}); ok {
			if lifecycleSettings, ok := indexSettings["lifecycle"].(map[string]interface{}); ok {
				actualPolicy, _ = lifecycleSettings["name"].(string)
			}
		}

		if actualPolicy == "" {
			return fmt.Errorf("component template %s has no index.lifecycle.name setting", name)
		}

		if actualPolicy != expectedPolicy {
			return fmt.Errorf("expected ILM policy %s, got %s", expectedPolicy, actualPolicy)
		}

		return nil
	}
}
