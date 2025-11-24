package securitylist_test

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

func TestAccResourceSecurityList(t *testing.T) {
	listID := "test-list-" + uuid.New().String()
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
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Test Security List"),
					"description": config.StringVariable("A test security list for IP addresses"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "Test Security List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "description", "A test security list for IP addresses"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "ip"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "created_by"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "updated_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_list.test", "updated_by"),
				),
			},
			{ // Update
				ConfigDirectory: acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Updated Security List"),
					"description": config.StringVariable("An updated test security list"),
					"type":        config.StringVariable("ip"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "name", "Updated Security List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "description", "An updated test security list"),
				),
			},
		},
	})
}

func TestAccResourceSecurityList_KeywordType(t *testing.T) {
	listID := "keyword-list-" + uuid.New().String()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			ensureListIndexExists(t)
		},
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: acctest.NamedTestCaseDirectory("keyword_type"),
				ConfigVariables: config.Variables{
					"list_id":     config.StringVariable(listID),
					"name":        config.StringVariable("Keyword Security List"),
					"description": config.StringVariable("A test security list for keywords"),
					"type":        config.StringVariable("keyword"),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_list.test", "type", "keyword"),
				),
			},
		},
	})
}
