package prebuilt_rules_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionPrebuiltRules = version.Must(version.NewVersion("8.0.0"))

func TestAccResourcePrebuiltRules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules),
				Config:   testAccPrebuiltRuleConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_install_prebuilt_rules.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_install_prebuilt_rules.test", "timelines_not_updated"),
				),
			},
		},
	})
}

func testAccPrebuiltRuleConfigBasic() string {
	return `
resource "elasticstack_kibana_install_prebuilt_rules" "test" {
  space_id = "default"
}
`
}
