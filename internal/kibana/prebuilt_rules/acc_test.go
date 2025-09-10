package prebuilt_rules_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr("elasticstack_kibana_prebuilt_rule.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "rules_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "rules_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "rules_not_updated"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "timelines_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "timelines_not_installed"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_prebuilt_rule.test", "timelines_not_updated"),
				),
			},
		},
	})
}

func TestAccResourcePrebuiltRule_withTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             testAccPrebuiltRuleDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minVersionPrebuiltRules),
				Config:   testAccPrebuiltRuleConfigWithTags(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_prebuilt_rule.test", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_prebuilt_rule.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_prebuilt_rule.test", "tags.0", "OS: Linux"),
					resource.TestCheckResourceAttr("elasticstack_kibana_prebuilt_rule.test", "tags.1", "OS: Windows"),
				),
			},
		},
	})
}

func testAccPrebuiltRuleDestroy(s *terraform.State) error {
	// For prebuilt rules, there's nothing to destroy
	// The rules remain in Kibana as they are managed by Elastic
	return nil
}

func testAccPrebuiltRuleConfigBasic() string {
	return `
resource "elasticstack_kibana_prebuilt_rule" "test" {
  space_id = "default"
}
`
}

func testAccPrebuiltRuleConfigWithTags() string {
	return `
resource "elasticstack_kibana_prebuilt_rule" "test" {
  space_id = "default"
  tags = ["OS: Linux", "OS: Windows"]
}
`
}
