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
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "entries.#"),
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
					resource.TestCheckResourceAttrSet(resourceName, "entries.#"),
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

func TestAccResourceExceptionItemEntryType_Match(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-match-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-match-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("match"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.value", "test-process"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_MatchAny(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-match-any-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-match-any-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("match_any"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.0", "process1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.1", "process2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.2", "process3"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_List(t *testing.T) {
	exceptionListID := fmt.Sprintf("test-exception-list-list-entry-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-list-entry-%s", uuid.New().String()[:8])
	valueListID := fmt.Sprintf("test-value-list-%s", uuid.New().String()[:8])
	valueListValue := "192.168.1.1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list"),
				ConfigVariables: config.Variables{
					"exception_list_id": config.StringVariable(exceptionListID),
					"item_id":           config.StringVariable(itemID),
					"value_list_id":     config.StringVariable(valueListID),
					"value_list_value":  config.StringVariable(valueListValue),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "source.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Exists(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-exists-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-exists-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("exists"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "file.hash.sha256"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Nested(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-nested-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-nested-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("nested"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "parent.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.field", "nested.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.value", "nested-value"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Wildcard(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-wildcard-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-wildcard-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "file.path"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.value", "/tmp/*.tmp"),
				),
			},
		},
	})
}
