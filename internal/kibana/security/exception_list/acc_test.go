package exception_list_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Exception lists were introduced in Kibana 7.9.0
var minExceptionListSupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionList(t *testing.T) {
	listName := "test-exception-list-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", listName),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Test exception list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "namespace_type", "single"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", listName+"-updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Updated exception list description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.#", "2"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListSupport),
				ConfigDirectory:          acctest.NamedTestCaseDirectory("updated"),
				ConfigVariables: config.Variables{
					"list_name": config.StringVariable(listName),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_security_exception_list.test",
			},
		},
	})
}
