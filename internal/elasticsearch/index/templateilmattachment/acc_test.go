package templateilmattachment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/templateilmattachment"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fleetSystemVersion is the system package version used by TestAccResourceIndexTemplateIlmAttachment_fleet.
// PreCheck installs it via Fleet API so it is available; must match version in testdata create/main.tf.
const fleetSystemVersion = "1.20.0"

const preservesTemplateIndexName = "logs-ilm-preserve"

func TestAccResourceIndexTemplateIlmAttachment_fleet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			// Skip before installing Fleet package if version is unsupported (PreCheck runs before SkipFunc).
			notSupported, err := versionutils.CheckIfVersionIsUnsupported(templateilmattachment.MinVersion)()
			if err != nil {
				t.Fatalf("checking version: %v", err)
			}
			if notSupported {
				t.Skip("Elasticsearch version does not support this test")
			}
			// Install system package via Fleet API so it is available (avoids conflict with tcp/sysmon_linux tests).
			client, err := clients.NewAcceptanceTestingClient()
			if err != nil {
				t.Fatalf("acceptance test client: %v", err)
			}
			fleetClient, err := client.GetFleetClient()
			if err != nil {
				t.Fatalf("Fleet client: %v", err)
			}
			diags := fleet.InstallPackage(context.Background(), fleetClient, "system", fleetSystemVersion, fleet.InstallPackageOptions{Force: true})
			if diags.HasError() {
				t.Fatalf("system package %s must be installable in Fleet registry: %v", fleetSystemVersion, diags)
			}
		},
		CheckDestroy: checkResourceDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(templateilmattachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable("test-fleet-policy-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"index_template", "logs-system.syslog"),
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"lifecycle_name", "test-fleet-policy-1"),
					checkComponentTemplateHasILM("logs-system.syslog@custom", "test-fleet-policy-1"),
				),
			},
			// Update
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(templateilmattachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"policy_name": config.StringVariable("test-fleet-policy-2"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"lifecycle_name", "test-fleet-policy-2"),
					checkComponentTemplateHasILM("logs-system.syslog@custom", "test-fleet-policy-2"),
				),
			},
			// Import
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(templateilmattachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
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

// TestAccResourceIndexTemplateIlmAttachment_preservesTemplateOnDestroy verifies that when the
// @custom component template has other settings, destroy only removes the ILM setting and leaves
// the template in place. PreCheck creates the template with index.number_of_shards; our resource
// adds ILM; after destroy the template must still exist without the ILM setting.
func TestAccResourceIndexTemplateIlmAttachment_preservesTemplateOnDestroy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createPreservesTestComponentTemplate(t)
		},
		CheckDestroy: checkPreservesTemplateDestroy,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(templateilmattachment.MinVersion),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"index_template": config.StringVariable(preservesTemplateIndexName),
					"policy_name":    config.StringVariable("test-preserves-policy"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"elasticstack_elasticsearch_index_template_ilm_attachment.test",
						"index_template", preservesTemplateIndexName),
					checkComponentTemplateHasILM(preservesTemplateIndexName+"@custom", "test-preserves-policy"),
				),
			},
		},
	})
}

func createPreservesTestComponentTemplate(t *testing.T) {
	t.Helper()
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("failed to create acceptance test client: %v", err)
	}
	name := preservesTemplateIndexName + "@custom"
	tpl := &models.ComponentTemplate{
		Name: name,
		Template: &models.Template{
			Settings: map[string]any{
				"index": map[string]any{
					"number_of_shards": "1",
				},
			},
		},
	}
	if diags := elasticsearch.PutComponentTemplate(context.Background(), client, tpl); diags.HasError() {
		t.Fatalf("failed to create component template %s: %v", name, diags)
	}
}

// checkPreservesTemplateDestroy verifies the template still exists and has no ILM setting, then
// deletes the template (cleanup of the fixture created in PreCheck).
func checkPreservesTemplateDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	name := preservesTemplateIndexName + "@custom"

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_elasticsearch_index_template_ilm_attachment" {
			continue
		}
		compID, sdkDiags := clients.CompositeIDFromStr(rs.Primary.ID)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to parse resource ID: %v", sdkDiags)
		}
		if compID.ResourceID != name {
			continue
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(ctx, client, name, true)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to get component template: %v", sdkDiags)
		}
		if tpl == nil {
			return fmt.Errorf("expected component template %s to still exist after destroy (only ILM should be removed)", name)
		}
		if tpl.ComponentTemplate.Template == nil {
			return fmt.Errorf("expected component template %s to still have a template section after destroy", name)
		}
		if tpl.ComponentTemplate.Template.Settings == nil {
			return fmt.Errorf("expected component template %s to still have settings after destroy (other settings should be preserved)", name)
		}
		if _, hasILM := tpl.ComponentTemplate.Template.Settings["index.lifecycle.name"]; hasILM {
			return fmt.Errorf("ILM setting still exists in component template %s", name)
		}
		// Verify the setting we created in PreCheck was preserved (only ILM should be removed)
		switch n := tpl.ComponentTemplate.Template.Settings["index.number_of_shards"].(type) {
		case string:
			if n != "1" {
				return fmt.Errorf("expected index.number_of_shards to be preserved as \"1\" after destroy, got %q", n)
			}
		case float64:
			if n != 1 {
				return fmt.Errorf("expected index.number_of_shards to be preserved as 1 after destroy, got %v", n)
			}
		default:
			got := tpl.ComponentTemplate.Template.Settings["index.number_of_shards"]
			return fmt.Errorf("expected index.number_of_shards to be preserved after destroy, got %v (type %T)", got, got)
		}

		// Cleanup: remove the fixture template created in PreCheck
		if diags := elasticsearch.DeleteComponentTemplate(ctx, client, name); diags.HasError() {
			return fmt.Errorf("failed to delete component template %s after check: %v", name, diags)
		}
		return nil
	}
	return nil
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

		compID, sdkDiags := clients.CompositeIDFromStr(rs.Primary.ID)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to parse resource ID: %v", sdkDiags)
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(context.Background(), client, compID.ResourceID, true)
		if sdkDiags.HasError() {
			return fmt.Errorf("failed to get component template: %v", sdkDiags)
		}

		// If the template still exists, check if ILM setting is removed
		if tpl != nil {
			if tpl.ComponentTemplate.Template != nil && tpl.ComponentTemplate.Template.Settings != nil {
				if _, hasILM := tpl.ComponentTemplate.Template.Settings["index.lifecycle.name"]; hasILM {
					return fmt.Errorf("ILM setting still exists in component template %s", compID.ResourceID)
				}
			}
		}
	}

	return nil
}

func checkComponentTemplateHasILM(name string, expectedPolicy string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		tpl, sdkDiags := elasticsearch.GetComponentTemplate(context.Background(), client, name, true)
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

		actualPolicy, _ := tpl.ComponentTemplate.Template.Settings["index.lifecycle.name"].(string)
		if actualPolicy == "" {
			return fmt.Errorf("component template %s has no index.lifecycle.name setting", name)
		}

		if actualPolicy != expectedPolicy {
			return fmt.Errorf("expected ILM policy %s, got %s", expectedPolicy, actualPolicy)
		}

		return nil
	}
}
