package exception_item_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceExceptionItem(t *testing.T) {
	listID := "test-exception-list-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	itemID := "test-exception-item-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "item_id", itemID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Test item description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "type", "simple"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "namespace_type", "single"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "space_id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "entries"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "item_id", itemID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Updated Exception Item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Updated item description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "os_types.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "os_types.0", "linux"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "entries"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemWithExpireTime(t *testing.T) {
	listID := "test-exception-list-expire-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	itemID := "test-exception-item-expire-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("expire_time"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "item_id", itemID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Exception Item with Expire Time"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "expire_time"),
				),
			},
		},
	})
}
