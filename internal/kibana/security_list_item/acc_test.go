package security_list_item_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceSecurityListItem(t *testing.T) {
	listID := "test-list-items-" + uuid.New().String()
	spaceID := "test-space-" + uuid.New().String()[:8]
	value1 := "test-value-1"
	valueUpdated := "test-value-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value1),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(valueUpdated),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", valueUpdated),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(valueUpdated),
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
	spaceID := "test-space-" + uuid.New().String()[:8]
	value := "test-value-with-meta"
	meta1 := `{"category":"suspicious","severity":"high"}`
	meta2 := `{"category":"malicious","notes":"Updated metadata","severity":"critical"}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create with meta
				ConfigDirectory: acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value),
					"meta":     config.StringVariable(meta1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "meta", meta1),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update meta
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value),
					"meta":     config.StringVariable(meta2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "meta", meta2),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value),
					"meta":     config.StringVariable(meta2),
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
	spaceName := "Test Security Lists Space"
	listName := "IP Blocklist"
	listType := "ip"
	value1 := "192.168.1.1"
	value2 := "10.0.0.1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create space, list, and list item
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_create"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value1),
				},
				Check: resource.ComposeTestCheckFunc(
					// Check space
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test", "name", spaceName),
					// Check list
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", listName),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", listType),
					// Check list item
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "space_id", spaceID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_id", listID),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value1),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
				),
			},
			{ // Update list item
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value2),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("space_update"),
				ConfigVariables: config.Variables{
					"space_id": config.StringVariable(spaceID),
					"list_id":  config.StringVariable(listID),
					"value":    config.StringVariable(value2),
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
	spaceID := "test-space-" + uuid.New().String()[:8]
	listItemID1 := "custom-item-id-1"
	listItemID2 := "custom-item-id-2"
	value1 := "test-value-1"
	value2 := "test-value-2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{ // Create with custom list_item_id
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_create"),
				ConfigVariables: config.Variables{
					"space_id":     config.StringVariable(spaceID),
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID1),
					"value":        config.StringVariable(value1),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_item_id", listItemID1),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value1),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list_item.test", "updated_by"),
				),
			},
			{ // Update list_item_id (should force replacement)
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_update"),
				ConfigVariables: config.Variables{
					"space_id":     config.StringVariable(spaceID),
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID2),
					"value":        config.StringVariable(value2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "list_item_id", listItemID2),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list_item.test", "value", value2),
				),
			},
			{ // Import
				ConfigDirectory: acctest.NamedTestCaseDirectory("with_list_item_id_update"),
				ConfigVariables: config.Variables{
					"space_id":     config.StringVariable(spaceID),
					"list_id":      config.StringVariable(listID),
					"list_item_id": config.StringVariable(listItemID2),
					"value":        config.StringVariable(value2),
				},
				ResourceName:      "elasticstack_kibana_security_list_item.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
