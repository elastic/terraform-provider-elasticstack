package import_saved_objects_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceImportSavedObjects(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "0"),
				),
			},
			{
				// Ensure a partially successful import doesn't throw a provider error
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("missing_ref"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_count", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "success_results.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_import_saved_objects.settings", "errors.#", "1"),
				),
			},
		},
	})
}
