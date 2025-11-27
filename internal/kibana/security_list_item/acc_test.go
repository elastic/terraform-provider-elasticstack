package security_list_item_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana_oapi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func ensureListIndexExists(t *testing.T) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	diags := kibana_oapi.CreateListIndex(context.Background(), kibanaClient, "default")
	if diags.HasError() {
		// It's OK if it already exists, we'll only fail on other errors
		for _, d := range diags {
			if d.Summary() != "Unexpected status code from server: got HTTP 409" {
				t.Fatalf("Failed to create list index: %v", d.Detail())
			}
		}
	}
}

func TestAccResourceSecurityListItem(t *testing.T) {
	listID := "test-list-items-" + uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			ensureListIndexExists(t)
		},
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-updated"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-updated"),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-updated"),
				},
				ResourceName:      "elasticstack_kibana_security_list_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityListItem_WithMeta(t *testing.T) {
	listID := "test-list-items-meta-" + uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			ensureListIndexExists(t)
		},
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create with meta
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-with-meta"),
					"meta":    config.StringVariable(`{"category":"suspicious","severity":"high"}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-with-meta"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "meta", `{"category":"suspicious","severity":"high"}`),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update meta
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-with-meta"),
					"meta":    config.StringVariable(`{"category":"malicious","notes":"Updated metadata","severity":"critical"}`),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-with-meta"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "meta", `{"category":"malicious","notes":"Updated metadata","severity":"critical"}`),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id": config.StringVariable(listID),
					"value":   config.StringVariable("test-value-with-meta"),
					"meta":    config.StringVariable(`{"category":"malicious","notes":"Updated metadata","severity":"critical"}`),
				},
				ResourceName:      "elasticstack_kibana_security_list_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityListItem_Space(t *testing.T) {
	spaceID := "test-space-" + uuid.New().String()
	listID := "test-list-" + uuid.New().String()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			ensureListIndexExistsInSpace(t, spaceID)
		},
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create space, list, and list item
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable("192.168.1.1"),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "name", "Test Security Lists Space"),
					// Check list
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "IP Blocklist"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "ip"),
					// Check list item
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "192.168.1.1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
				),
			},
			{ // Update list item
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable("10.0.0.1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "10.0.0.1"),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable("10.0.0.1"),
				},
				ResourceName:      "elasticstack_kibana_security_list_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceSecurityListItem_WithListItemID(t *testing.T) {
	listID := "test-list-items-with-id-" + uuid.New().String()
	listItemID1 := "custom-item-id-1"
	listItemID2 := "custom-item-id-2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			ensureListIndexExists(t)
		},
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create with custom list_item_id
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_create"),
				ConfigVariables: config.Variables{
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID1),
					"value":        config.StringVariable("test-value-1"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_item_id", listItemID1),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update list_item_id (should force replacement)
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_update"),
				ConfigVariables: config.Variables{
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID2),
					"value":        config.StringVariable("test-value-2"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_item_id", listItemID2),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", "test-value-2"),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_update"),
				ConfigVariables: config.Variables{
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID2),
					"value":        config.StringVariable("test-value-2"),
				},
				ResourceName:      "elasticstack_kibana_security_list_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func ensureListIndexExistsInSpace(t *testing.T, spaceID string) {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	kibanaClient, err := client.GetKibanaOapiClient()
	if err != nil {
		t.Fatalf("Failed to get Kibana client: %v", err)
	}

	diags := kibana_oapi.CreateListIndex(context.Background(), kibanaClient, spaceID)
	if diags.HasError() {
		// It's OK if it already exists, we'll only fail on other errors
		for _, d := range diags {
			if d.Summary() != "Unexpected status code from server: got HTTP 409" {
				t.Fatalf("Failed to create list index in space %s: %v", spaceID, d.Detail())
			}
		}
	}
}
