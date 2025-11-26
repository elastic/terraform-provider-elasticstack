package security_exception_list_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minExceptionListAPISupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list"),
					"name":           config.StringVariable("Test Exception List"),
					"description":    config.StringVariable("Test exception list for acceptance tests"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "list_id", "test-exception-list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Test exception list for acceptance tests"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "namespace_type", "single"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.0", "test"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_by"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list"),
					"name":           config.StringVariable("Test Exception List Updated"),
					"description":    config.StringVariable("Updated description"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.1", "updated"),
				),
			},
		},
	})
}

func TestAccResourceExceptionListWithSpace(t *testing.T) {
	resourceName := "elasticstack_kibana_security_exception_list.test"
	spaceResourceName := "elasticstack_kibana_space.test"
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"list_id":        config.StringVariable("test-exception-list-space"),
					"name":           config.StringVariable("Test Exception List in Space"),
					"description":    config.StringVariable("Test exception list in custom space"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("space")),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Exception Lists"),

					// Check exception list attributes
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "list_id", "test-exception-list-space"),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Exception List in Space"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test exception list in custom space"),
					resource.TestCheckResourceAttr(resourceName, "type", "detection"),
					resource.TestCheckResourceAttr(resourceName, "namespace_type", "single"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "space"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"list_id":        config.StringVariable("test-exception-list-space"),
					"name":           config.StringVariable("Test Exception List in Space Updated"),
					"description":    config.StringVariable("Updated description in space"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("space"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes remain the same
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Exception Lists"),

					// Check updated exception list attributes
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Exception List in Space Updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description in space"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "space"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "updated"),
				),
			},
		},
	})
}
