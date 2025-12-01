package security_exception_item_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

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

var minExceptionItemAPISupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionItem(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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
			{ // Import
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
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "updated"),
				),
			},
			{ // Import
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
				ResourceName:      resourceName,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
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

func TestAccResourceExceptionItemValidation(t *testing.T) {
	listID := fmt.Sprintf("test-exception-list-validation-%s", uuid.New().String()[:8])
	itemID := fmt.Sprintf("test-exception-item-validation-%s", uuid.New().String()[:8])

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			// Test 1: Match entry missing value
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: checkResourceExceptionItemDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_item.test", "expire_time", "2025-12-31T23:59:59Z"),
				),
			},
			{
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionItemAPISupport),
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
