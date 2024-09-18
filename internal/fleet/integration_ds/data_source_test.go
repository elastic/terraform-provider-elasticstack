package integration_ds_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var minVersionIntegrationDataSource = version.Must(version.NewVersion("8.6.0"))

func TestAccDataSourceIntegration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionIntegrationDataSource),
				Config:   testAccDataSourceIntegration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_integration.test", "name", "tcp"),
					checkResourceAttrStringNotEmpty("data.elasticstack_fleet_integration.test", "version"),
				),
			},
		},
	})
}

const testAccDataSourceIntegration = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

data "elasticstack_fleet_integration" "test" {
  name = "tcp"
}
`

// checkResourceAttrStringNotEmpty verifies that the string value at key
// is not empty.
func checkResourceAttrStringNotEmpty(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s in %s", name, ms.Path)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("no primary instance: %s in %s", name, ms.Path)
		}

		v, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}
		if v == "" {
			return fmt.Errorf("%s: Attribute '%s' expected non-empty string", name, key)
		}

		return nil
	}
}
