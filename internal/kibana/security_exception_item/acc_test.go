package security_exception_item_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var MinVersionExpireTime = version.Must(version.NewVersion("8.7.2"))
var MinExceptionItemVersion = version.Must(version.NewVersion("7.9.0"))

// https://github.com/elastic/kibana/pull/159223
// These versions don't respect item_id which breaks most tests
const exceptionItemBugVersionConstraint = "!=8.7.0,!=8.7.1"
const minExceptionItemAPISupportConstraint = ">=7.9.0"

var allTestsVersionsConstraint, _ = version.NewConstraint(exceptionItemBugVersionConstraint + "," + minExceptionItemAPISupportConstraint)

func TestAccResourceExceptionItem(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{ // Import
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
				ResourceName:      "elasticstack_kibana_security_exception_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceExceptionItem_BasicUsage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(MinExceptionItemVersion),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_create"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-basic"),
					"name":           config.StringVariable("Test Exception Item - Basic"),
					"description":    config.StringVariable("Test exception item without explicit item_id"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "item_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item - Basic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Test exception item without explicit item_id"),
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
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-basic"),
					"name":           config.StringVariable("Test Exception Item - Basic Updated"),
					"description":    config.StringVariable("Updated basic exception item"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "item_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item - Basic Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Updated basic exception item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.1", "updated"),
				),
			},
			{ // Import
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("basic_update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-basic"),
					"name":           config.StringVariable("Test Exception Item - Basic Updated"),
					"description":    config.StringVariable("Updated basic exception item"),
					"type":           config.StringVariable("simple"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.ListVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				ResourceName:      "elasticstack_kibana_security_exception_item.test",
				ImportState:       true,
				ImportStateVerify: true,
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
		CheckDestroy:             checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "entries.#"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "updated"),
				),
			},
			{ // Import
				SkipFunc:        versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceExceptionItemNamespaceType_Agnostic(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-agnostic-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-agnostic-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agnostic_create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "item_id", itemID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item - Agnostic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Test exception item with agnostic namespace type"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "type", "simple"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "namespace_type", "agnostic"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_item.test", "entries.#"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agnostic_update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "name", "Test Exception Item - Agnostic Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "description", "Updated agnostic exception item"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "namespace_type", "agnostic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "updated"),
				),
			},
			{ // Import
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("agnostic_update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ResourceName:      "elasticstack_kibana_security_exception_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Match(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-match-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-match-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("match_multiple"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.value", "test-process"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "user.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.value", "test-user"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "host.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.value", "excluded-host"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_MatchAny(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-match-any-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-match-any-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("match_any_multiple"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "user.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.values.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "file.extension"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.values.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_List(t *testing.T) {
	exceptionListID := fmt.Sprintf("test-exception-list-list-entry-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-list-entry-%s", uuid.New().String()[:8])
	valueListIDIP := fmt.Sprintf("test-value-list-ip-%s", uuid.New().String()[:8])
	valueListIDKeyword := fmt.Sprintf("test-value-list-keyword-%s", uuid.New().String()[:8])
	valueListIDIPRange := fmt.Sprintf("test-value-list-ip-range-%s", uuid.New().String()[:8])
	spaceID := fmt.Sprintf("test-space-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list"),
				ConfigVariables: config.Variables{
					"space_id":          config.StringVariable(spaceID),
					"exception_list_id": config.StringVariable(exceptionListID),
					"item_id":           config.StringVariable(itemID),
					"value_list_id":     config.StringVariable(valueListIDIP),
					"value_list_value":  config.StringVariable("192.168.1.1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "source.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDIP),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_update"),
				ConfigVariables: config.Variables{
					"space_id":           config.StringVariable(spaceID),
					"exception_list_id":  config.StringVariable(exceptionListID),
					"item_id":            config.StringVariable(itemID),
					"value_list_id":      config.StringVariable(valueListIDIP),
					"value_list_value":   config.StringVariable("192.168.1.1"),
					"value_list_value_2": config.StringVariable("10.0.0.1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "source.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDIP),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_keyword"),
				ConfigVariables: config.Variables{
					"space_id":          config.StringVariable(spaceID),
					"exception_list_id": config.StringVariable(exceptionListID),
					"item_id":           config.StringVariable(itemID),
					"value_list_id":     config.StringVariable(valueListIDKeyword),
					"value_list_value":  config.StringVariable("test-process"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDKeyword),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "keyword"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_keyword_update"),
				ConfigVariables: config.Variables{
					"space_id":           config.StringVariable(spaceID),
					"exception_list_id":  config.StringVariable(exceptionListID),
					"item_id":            config.StringVariable(itemID),
					"value_list_id":      config.StringVariable(valueListIDKeyword),
					"value_list_value":   config.StringVariable("test-process"),
					"value_list_value_2": config.StringVariable("another-process"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDKeyword),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "keyword"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_ip_range"),
				ConfigVariables: config.Variables{
					"space_id":          config.StringVariable(spaceID),
					"exception_list_id": config.StringVariable(exceptionListID),
					"item_id":           config.StringVariable(itemID),
					"value_list_id":     config.StringVariable(valueListIDIPRange),
					"value_list_value":  config.StringVariable("192.168.1.0/24"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDIPRange),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip_range"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_ip_range_update"),
				ConfigVariables: config.Variables{
					"space_id":           config.StringVariable(spaceID),
					"exception_list_id":  config.StringVariable(exceptionListID),
					"item_id":            config.StringVariable(itemID),
					"value_list_id":      config.StringVariable(valueListIDIPRange),
					"value_list_value":   config.StringVariable("192.168.1.0/24"),
					"value_list_value_2": config.StringVariable("10.0.0.0/16"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "destination.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDIPRange),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip_range"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("list_multiple"),
				ConfigVariables: config.Variables{
					"space_id":                 config.StringVariable(spaceID),
					"exception_list_id":        config.StringVariable(exceptionListID),
					"item_id":                  config.StringVariable(itemID),
					"value_list_id_ip":         config.StringVariable(valueListIDIP),
					"value_list_id_keyword":    config.StringVariable(valueListIDKeyword),
					"value_list_value_ip":      config.StringVariable("192.168.1.1"),
					"value_list_value_keyword": config.StringVariable("test-process"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "source.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.id", valueListIDIP),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.list.type", "ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.list.id", valueListIDKeyword),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.list.type", "keyword"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "destination.ip"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.list.id", valueListIDIP),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.list.type", "ip"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Exists(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-exists-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-exists-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("exists_multiple"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "file.hash.sha256"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "process.code_signature.trusted"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "network.protocol"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Nested(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-nested-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-nested-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_match_any"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "parent.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.field", "nested.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.values.0", "value1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.values.1", "value2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.values.2", "value3"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_exists"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "parent.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.field", "nested.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.operator", "included"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("nested_multiple"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "parent.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.entries.0.field", "nested.field"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "process.parent"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.entries.0.field", "process.parent.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "nested"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "file.attributes"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.entries.0.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.entries.0.field", "file.attributes.hidden"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemEntryType_Wildcard(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-wildcard-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-wildcard-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
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
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("wildcard_multiple"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "file.path"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.value", "/tmp/*.tmp"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "process.command_line"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.value", "*powershell*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "dns.question.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.value", "*.malicious.com"),
				),
			},
		},
	})
}

func TestAccResourceExceptionItemValidation(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-validation-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-validation-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			// Test 1: Match entry missing value
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_match_missing_value"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'match' requires 'value' to be set"),
				PlanOnly:    true,
			},
			// Test 2: Match entry missing operator
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_match_missing_operator"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'match' requires 'operator' to be set"),
				PlanOnly:    true,
			},
			// Test 3: Wildcard entry missing value
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_wildcard_missing_value"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'wildcard' requires 'value' to be set"),
				PlanOnly:    true,
			},
			// Test 4: MatchAny entry missing values
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_match_any_missing_values"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'match_any' requires 'values' to be set"),
				PlanOnly:    true,
			},
			// Test 5: MatchAny entry missing operator
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_match_any_missing_operator"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'match_any' requires 'operator' to be set"),
				PlanOnly:    true,
			},
			// Test 6: List entry missing list object
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_list_missing_list_object"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'list' requires 'list' object to be set"),
				PlanOnly:    true,
			},
			// Test 7: List entry missing list.id
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_list_missing_list_id"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile(`attribute "id" is required`),
				PlanOnly:    true,
			},
			// Test 8: List entry missing list.type
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_list_missing_list_type"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile(`attribute "type" is required`),
				PlanOnly:    true,
			},
			// Test 9: Exists entry missing operator
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_exists_missing_operator"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'exists' requires 'operator' to be set"),
				PlanOnly:    true,
			},
			// Test 10: Nested entry missing entries
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_nested_missing_entries"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Entry type 'nested' requires 'entries' to be set"),
				PlanOnly:    true,
			},
			// Test 11: Nested entry with invalid entry type
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_nested_invalid_entry_type"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile(`(Nested entry .* has invalid type|value must be one of:.*"match".*"match_any".*"exists")`),
				PlanOnly:    true,
			},
			// Test 12: Nested match entry missing value
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_nested_entry_missing_value"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile("Nested entry type 'match' requires 'value' to be set"),
				PlanOnly:    true,
			},
			// Test 13: Nested entry missing operator
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("validation_nested_entry_missing_operator"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				ExpectError: regexp.MustCompile(`(Nested entry requires 'operator' to be set|attribute "operator" is required)`),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccResourceExceptionItem_Complex(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-complex-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-complex-%s", uuid.New().String()[:8])

	// Generate an expiration time 2 days from now with milliseconds set to 0
	// since default go time formatting may truncate milliseconds in date serialization
	// resulting in 4xx responses from Kibana
	expireTime := time.Now().AddDate(0, 0, 2).UTC().Truncate(24 * time.Hour).Format("2006-01-02T15:04:05.000Z")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complex_create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "os_types.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "linux"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "macos"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "complex"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complex_update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "os_types.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "linux"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "macos"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "windows"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "complex"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "updated"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complex_replace_entries"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"item_id": config.StringVariable(itemID),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.field", "file.path"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.#", "3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.0", "/usr/bin/*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.1", "/usr/sbin/*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.0.values.2", "/bin/*"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.field", "file.hash.sha256"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.1.operator", "included"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.field", "process.command_line"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.operator", "excluded"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "entries.2.value", "*malicious*"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionMeetsConstraints(allTestsVersionsConstraint),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complex_multiple_items"),
				ConfigVariables: config.Variables{
					"list_id":   config.StringVariable(listID),
					"item_id_1": config.StringVariable(itemID + "-1"),
					"item_id_2": config.StringVariable(itemID + "-2"),
					"item_id_3": config.StringVariable(itemID + "-3"),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check first item
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "item_id", itemID+"-1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "name", "Test Exception Item 1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "entries.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "entries.0.type", "match"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "entries.0.field", "process.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "entries.0.value", "process1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test1", "os_types.#", "1"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test1", "os_types.*", "linux"),

					// Check second item
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "item_id", itemID+"-2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "name", "Test Exception Item 2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "entries.#", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "entries.0.type", "match_any"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "entries.0.field", "user.name"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "entries.0.values.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test2", "os_types.#", "2"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test2", "os_types.*", "linux"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test2", "os_types.*", "macos"),

					// Check third item
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "item_id", itemID+"-3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "name", "Test Exception Item 3"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "entries.#", "2"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "entries.0.type", "wildcard"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "entries.0.field", "file.path"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "entries.1.type", "exists"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "entries.1.field", "file.hash.sha256"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test3", "os_types.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test3", "os_types.*", "linux"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test3", "os_types.*", "macos"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test3", "os_types.*", "windows"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(MinVersionExpireTime),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("complex_update_expire_time"),
				ConfigVariables: config.Variables{
					"list_id":     config.StringVariable(listID),
					"item_id":     config.StringVariable(itemID),
					"expire_time": config.StringVariable(expireTime),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "os_types.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "linux"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "macos"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "os_types.*", "windows"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "complex"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_item.test", "tags.*", "updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "expire_time", expireTime),
				),
			},
		},
	})
}

func checkResourceExceptionItemDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_exception_item" {
			continue
		}

		compId, compDiags := clients.CompositeIdFromStr(rs.Primary.ID)
		if compDiags.HasError() {
			return diagutil.SdkDiagsAsError(compDiags)
		}

		// Try to read the exception item
		id := kbapi.SecurityExceptionsAPIExceptionListItemId(compId.ResourceId)
		params := &kbapi.ReadExceptionListItemParams{
			Id: &id,
		}

		// If namespace_type is available in the state, use it
		if nsType, ok := rs.Primary.Attributes["namespace_type"]; ok && nsType != "" {
			nsTypeVal := kbapi.SecurityExceptionsAPIExceptionNamespaceType(nsType)
			params.NamespaceType = &nsTypeVal
		}

		item, diags := kibana_oapi.GetExceptionListItem(context.Background(), oapiClient, compId.ClusterId, params)
		if diags.HasError() {
			// If we get an error, it might be that the resource doesn't exist, which is what we want
			continue
		}

		if item != nil {
			return fmt.Errorf("Exception item (%s) still exists in space (%s)", compId.ResourceId, compId.ClusterId)
		}
	}
	return nil
}
