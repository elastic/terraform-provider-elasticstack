package enrich_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceEnrichPolicyFW(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkEnrichPolicyDestroyFW(name),
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichPolicyFW(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "match_field", `email`),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "indices.0", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.0", "first_name"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "enrich_fields.1", "last_name"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "query", "{\"match_all\":{}}"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceEnrichPolicyFW(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccEnrichPolicyDataSourceFW(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "name", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "policy_type", "match"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "match_field", "email"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "indices.0", name),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.0", "first_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "enrich_fields.1", "last_name"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_enrich_policy.test", "query", "{\"match_all\":{}}"),
				),
			},
		},
	})
}

func TestAccResourceEnrichPolicyFromSDK(t *testing.T) {
	name := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the enrich policy with the last provider version where the enrich policy resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.15",
					},
				},
				Config: testAccEnrichPolicyFW(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccEnrichPolicyFW(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "name", name),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "policy_type", "match"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_enrich_policy.policy", "execute", "true"),
				),
			},
		},
	})
}

func testAccEnrichPolicyFW(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      email      = { type = "text" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = "%s"
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
	query = <<-EOD
	{"match_all": {}}
	EOD
}
	`, name, name)
}

func testAccEnrichPolicyDataSourceFW(name string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
  name = "%s"

  mappings = jsonencode({
    properties = {
      email      = { type = "text" }
      first_name = { type = "text" }
      last_name  = { type = "text" }
    }
  })
  deletion_protection = false
}

resource "elasticstack_elasticsearch_enrich_policy" "policy" {
  name          = "%s"
  policy_type   = "match"
  indices       = [elasticstack_elasticsearch_index.my_index.name]
  match_field   = "email"
  enrich_fields = ["first_name", "last_name"]
	query = <<-EOD
	{"match_all": {}}
	EOD
}

data "elasticstack_elasticsearch_enrich_policy" "test" {
	name = elasticstack_elasticsearch_enrich_policy.policy.name
}
	`, name, name)
}

func checkEnrichPolicyDestroyFW(name string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		client, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return err
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "elasticstack_elasticsearch_enrich_policy" {
				continue
			}
			compId, _ := clients.CompositeIdFromStr(rs.Primary.ID)
			if compId.ResourceId != name {
				return fmt.Errorf("Found unexpectedly enrich policy: %s", compId.ResourceId)
			}
			esClient, err := client.GetESClient()
			if err != nil {
				return err
			}
			req := esClient.EnrichGetPolicy.WithName(compId.ResourceId)
			res, err := esClient.EnrichGetPolicy(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode == http.StatusFound {
				var policiesResponse map[string]any
				if err := json.NewDecoder(res.Body).Decode(&policiesResponse); err != nil {
					return err
				}
				if len(policiesResponse["policies"].([]any)) != 0 {
					return fmt.Errorf("Enrich policy (%s) still exists", compId.ResourceId)
				}
			}
		}
		return nil
	}
}
