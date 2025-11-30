package security_exception_list_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					"tags":           config.SetVariable(config.StringVariable("test")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "list_id", "test-exception-list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Test exception list for acceptance tests"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "namespace_type", "single"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "test"),
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
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Updated description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "updated"),
				),
			},
			{ // Import
				SkipFunc:                 versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ProtoV6ProviderFactories: acctest.Providers,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list"),
					"name":           config.StringVariable("Test Exception List Updated"),
					"description":    config.StringVariable("Updated description"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("updated")),
				},
				ResourceName:      "elasticstack_kibana_security_exception_list.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceExceptionListAgnostic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		CheckDestroy:             checkResourceExceptionListDestroy,
		Steps: []resource.TestStep{
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("agnostic_create"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-agnostic"),
					"name":           config.StringVariable("Test Exception List Agnostic"),
					"description":    config.StringVariable("Test agnostic exception list for acceptance tests"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("agnostic"),
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("agnostic")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "list_id", "test-exception-list-agnostic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List Agnostic"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Test agnostic exception list for acceptance tests"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "namespace_type", "agnostic"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "agnostic"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_by"),
				),
			},
			{
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("agnostic_update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-agnostic"),
					"name":           config.StringVariable("Test Exception List Agnostic Updated"),
					"description":    config.StringVariable("Updated agnostic description"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("agnostic"),
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("agnostic"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List Agnostic Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Updated agnostic description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "agnostic"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_security_exception_list.test", "tags.*", "updated"),
				),
			},
			{ // Import
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("agnostic_update"),
				ConfigVariables: config.Variables{
					"list_id":        config.StringVariable("test-exception-list-agnostic"),
					"name":           config.StringVariable("Test Exception List Agnostic Updated"),
					"description":    config.StringVariable("Updated agnostic description"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("agnostic"),
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("agnostic"), config.StringVariable("updated")),
				},
				ResourceName:      "elasticstack_kibana_security_exception_list.test",
				ImportState:       true,
				ImportStateVerify: true,
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
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("space")),
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
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
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
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("space"), config.StringVariable("updated")),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space attributes remain the same
					resource.TestCheckResourceAttr(spaceResourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(spaceResourceName, "name", "Test Space for Exception Lists"),

					// Check updated exception list attributes
					resource.TestCheckResourceAttr(resourceName, "space_id", spaceID),
					resource.TestCheckResourceAttr(resourceName, "name", "Test Exception List in Space Updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description in space"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "space"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "updated"),
				),
			},
			{ // Import
				SkipFunc:        versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id":       config.StringVariable(spaceID),
					"list_id":        config.StringVariable("test-exception-list-space"),
					"name":           config.StringVariable("Test Exception List in Space Updated"),
					"description":    config.StringVariable("Updated description in space"),
					"type":           config.StringVariable("detection"),
					"namespace_type": config.StringVariable("single"),
					"tags":           config.SetVariable(config.StringVariable("test"), config.StringVariable("space"), config.StringVariable("updated")),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceExceptionListDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_security_exception_list" {
			continue
		}

		compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)

		// Try to read the exception list with its namespace_type
		id := kbapi.SecurityExceptionsAPIExceptionListId(compId.ResourceId)
		params := &kbapi.ReadExceptionListParams{
			Id: &id,
		}

		// If namespace_type is available in the state, use it
		if nsType, ok := rs.Primary.Attributes["namespace_type"]; ok && nsType != "" {
			nsTypeVal := kbapi.SecurityExceptionsAPIExceptionNamespaceType(nsType)
			params.NamespaceType = &nsTypeVal
		}

		list, diags := kibana_oapi.GetExceptionList(context.Background(), oapiClient, compId.ClusterId, params)
		if diags.HasError() {
			// If we get an error, it might be that the resource doesn't exist, which is what we want
			continue
		}

		if list != nil {
			return fmt.Errorf("Exception list (%s) still exists in space (%s)", compId.ResourceId, compId.ClusterId)
		}
	}
	return nil
}
