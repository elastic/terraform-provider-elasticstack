package exception_item_test

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

var minExceptionItemAPISupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionItem(t *testing.T) {
	entriesCreate := `[{"field":"process.name","operator":"included","type":"match","value":"test-process"}]`
	entriesUpdate := `[{"field":"process.name","operator":"included","type":"match","value":"test-process-updated"}]`

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-for-item"),
					"item_id":        config.StringVariable("test-exception-item"),
					"name":           config.StringVariable("Test Exception Item"),
					"description":    config.StringVariable("Test exception item for acceptance tests"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"entries":        config.StringVariable(entriesCreate),
					"tags":           config.ListVariable(config.StringVariable("test")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "item_id", "test-exception-item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Test exception item for acceptance tests"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "type", "simple"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "namespace_type", "single"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.0", "test"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "entries"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "created_by"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-for-item"),
					"item_id":        config.StringVariable("test-exception-item"),
					"name":           config.StringVariable("Test Exception Item Updated"),
					"description":    config.StringVariable("Updated description"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"entries":        config.StringVariable(entriesUpdate),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.1", "updated"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemWithSpace(t *testing.T) {
	resourceName := "elasticstack_kibana_security_exception_item.test"
	spaceResourceName := "elasticstack_kibana_space.test"
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])
	entriesCreate := `[{"field":"process.name","operator":"included","type":"match","value":"test-process-space"}]`
	entriesUpdate := `[{"field":"process.name","operator":"included","type":"match","value":"test-process-space-updated"}]`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"list_id":        config.StringVariable("test-exception-list-for-item-space"),
					"item_id":        config.StringVariable("test-exception-item-space"),
					"name":           config.StringVariable("Test Exception Item in Space"),
					"description":    config.StringVariable("Test exception item in custom space"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"entries":        config.StringVariable(entriesCreate),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("space")),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Exception Items"),

					// Check exception item attributes
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "item_id", "test-exception-item-space"),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Exception Item in Space"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test exception item in custom space"),
					resource.TestCheckResourceAttr(resourceName, "type", "simple"),
					resource.TestCheckResourceAttr(resourceName, "namespace_type", "single"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "space"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "entries"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"list_id":        config.StringVariable("test-exception-list-for-item-space"),
					"item_id":        config.StringVariable("test-exception-item-space"),
					"name":           config.StringVariable("Test Exception Item in Space Updated"),
					"description":    config.StringVariable("Updated description in space"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"entries":        config.StringVariable(entriesUpdate),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("space"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes remain the same
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Exception Items"),

					// Check updated exception item attributes
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Exception Item in Space Updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description in space"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "space"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "updated"),
				),
			},
		},
	})
}
