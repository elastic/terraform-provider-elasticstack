package acctest

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/examples"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccExamples verifies that all example files in examples/data-sources and examples/resources are valid and can be applied.
func TestAccExamples(t *testing.T) {
	// Test data source examples
	testExamplesInFS(t, examples.DataSources)

	// Test resource examples
	testExamplesInFS(t, examples.Resources)
}

func testExamplesInFS(t *testing.T, examplesFS embed.FS) {
	err := fs.WalkDir(examplesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process .tf files, skip import.sh files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".tf") {
			return nil
		}

		// Read the file content
		content, err := fs.ReadFile(examplesFS, path)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", path, err)
			return nil
		}

		t.Run(path, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				PreCheck: func() { PreCheck(t) },
				Steps: []resource.TestStep{
					{
						ProtoV6ProviderFactories: Providers,
						Config:                   string(content),
						Check: resource.ComposeTestCheckFunc(
							// Basic check to ensure the config was applied
							func(s *terraform.State) error {
								if len(s.RootModule().Resources) == 0 {
									return fmt.Errorf("no resources found in state")
								}
								return nil
							},
						),
					},
				},
			})
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk example files %v", err)
	}
}
