package parameter_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	providerConfig = `
provider "elasticstack" {
	kibana {}
}
`
)

var (
	minKibanaParameterAPIVersion = version.Must(version.NewVersion("8.12.0"))
)

func TestSyntheticParameterResource(t *testing.T) {
	resourceId := "elasticstack_kibana_synthetics_parameter.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key"
	value = "test-value"
	description = "Test description"
	tags = ["a", "b"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceId, "key", "test-key"),
					resource.TestCheckResourceAttr(resourceId, "value", "test-value"),
					resource.TestCheckResourceAttr(resourceId, "description", "Test description"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "a"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "b"),
				),
			},
			// ImportState testing
			{
				SkipFunc:          versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key"
	value = "test-value"
	description = "Test description"
	tags = ["a", "b"]
}
`,
			},
			// Update and Read testing
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minKibanaParameterAPIVersion),
				Config: providerConfig + `
resource "elasticstack_kibana_synthetics_parameter" "test" {
	key = "test-key-2"
	value = "test-value-2"
	description = "Test description 2"
	tags = ["c", "d", "e"]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceId, "key", "test-key-2"),
					resource.TestCheckResourceAttr(resourceId, "value", "test-value-2"),
					resource.TestCheckResourceAttr(resourceId, "description", "Test description 2"),
					resource.TestCheckResourceAttr(resourceId, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceId, "tags.0", "c"),
					resource.TestCheckResourceAttr(resourceId, "tags.1", "d"),
					resource.TestCheckResourceAttr(resourceId, "tags.2", "e"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
