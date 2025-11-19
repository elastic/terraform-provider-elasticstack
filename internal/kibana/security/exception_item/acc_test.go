package exception_item_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Exception items were introduced in Kibana 7.9.0
var minExceptionItemSupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionItem(t *testing.T) {
	listName := "test-exception-list-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	itemName := "test-exception-item-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
					"item_name": config.StringVariable(itemName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", itemName),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Test exception item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "type", "simple"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
					"item_name": config.StringVariable(itemName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", itemName+"-updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Updated exception item description"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
					"item_name": config.StringVariable(itemName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_security_exception_item.test",
			},
		},
	})
}
