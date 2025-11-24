package securitylistitem_test

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
		},
	})
}
