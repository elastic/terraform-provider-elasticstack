package system_user_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest/checks"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceSecuritySystemUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecuritySystemUserCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", "remote_monitoring_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
			{
				Config: testAccResourceSecuritySystemUserUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", "remote_monitoring_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
					checks.CheckUserCanAuthenticate("remote_monitoring_user", "new_password"),
				),
			},
		},
	})
}

func TestAccResourceSecuritySystemUserNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceSecuritySystemUserNotFound,
				ExpectError: regexp.MustCompile(`System user "not_system_user" not found`),
			},
		},
	})
}

func TestAccResourceSecuritySystemUserFromSDK(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Create the system user with the last provider version where the system user resource was built on the SDK
				ExternalProviders: map[string]resource.ExternalProvider{
					"elasticstack": {
						Source:            "elastic/elasticstack",
						VersionConstraint: "0.11.15",
					},
				},
				Config: testAccResourceSecuritySystemUserCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", "remote_monitoring_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				Config:                   testAccResourceSecuritySystemUserCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "username", "remote_monitoring_user"),
					resource.TestCheckResourceAttr("elasticstack_elasticsearch_security_system_user.remote_monitoring_user", "enabled", "true"),
				),
			},
		},
	})
}

const testAccResourceSecuritySystemUserCreate = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "remote_monitoring_user" {
  username  = "remote_monitoring_user"
}
`

const testAccResourceSecuritySystemUserUpdate = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "remote_monitoring_user" {
  username  = "remote_monitoring_user"
  password  = "new_password"
}
	`
const testAccResourceSecuritySystemUserNotFound = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "test" {
  username  = "not_system_user"
  password  = "new_password"
}
	`
