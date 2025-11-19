package exception_container_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceExceptionContainer(t *testing.T) {
	listID := "test-exception-list-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "name", "Test Exception Container"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "description", "Test description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "namespace_type", "single"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "space_id", "default"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "name", "Updated Exception Container"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "tags.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "tags.0", "tag1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "tags.1", "tag2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "os_types.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "os_types.0", "linux"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "os_types.1", "windows"),
				),
			},
		},
	})
}

func TestAccResourceExceptionContainerWithNamespaceType(t *testing.T) {
	listID := "test-exception-list-agnostic-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agnostic"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "name", "Agnostic Exception Container"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_container.test", "namespace_type", "agnostic"),
				),
			},
		},
	})
}
