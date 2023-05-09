package fleet_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var minVersionEnrollmentTokens = version.Must(version.NewVersion("8.0.0"))

func TestAccDataSourceEnrollmentTokens(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionEnrollmentTokens),
				Config:   testAccDataSourceEnrollmentToken,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_fleet_enrollment_tokens.test", "policy_id", "223b1bf8-240f-463f-8466-5062670d0754"),
				),
			},
		},
	})
}

const testAccDataSourceEnrollmentToken = `
provider "elasticstack" {
  kibana {}
}

data "elasticstack_fleet_enrollment_tokens" "test" {
	policy_id = "223b1bf8-240f-463f-8466-5062670d0754"
}
`
