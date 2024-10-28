package security_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSecurityRole(t *testing.T) {
	minSupportedRemoteIndicesVersion := version.Must(version.NewSemver("8.10.0"))
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityRole,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
				),
			},
			{
				Config:   testAccDataSourceSecurityRoleRemoteIndices,
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedRemoteIndicesVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.*.clusters.*", "test-cluster2"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "remote_indices.*.names.*", "sample2"),
				),
			},
			{
				Config:   testAccDataSourceSecurityRoleWithDescription,
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minSupportedDescriptionVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "name", "data_source_test"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "cluster.*", "all"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.names", []string{"index1", "index2"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.privileges.*", "all"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "indices.0.allow_restricted_indices", "true"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.application", "myapp"),
					utils.TestCheckResourceListAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.privileges", []string{"admin", "read"}),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "applications.0.resources.*", "*"),
					resource.TestCheckTypeSetElemAttr("data.elasticstack_elasticsearch_security_role.test", "run_as.*", "other_user"),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "metadata", `{"version":1}`),
					resource.TestCheckResourceAttr("data.elasticstack_elasticsearch_security_role.test", "description", `Test data source`),
				),
			},
		},
	})
}

const testAccDataSourceSecurityRole = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "data_source_test"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
`

const testAccDataSourceSecurityRoleWithDescription = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "data_source_test"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })

  description =  "Test data source"
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
`

const testAccDataSourceSecurityRoleRemoteIndices = `
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "test" {
  name    = "data_source_test"
  cluster = ["all"]

  indices {
    names                    = ["index1", "index2"]
    privileges               = ["all"]
    allow_restricted_indices = true
  }

  remote_indices {
	clusters = ["test-cluster2"]
	field_security {
	  grant = ["sample"]
	  except = []
	}
	names = ["sample2"]
	privileges = ["create", "read", "write"]
  }

  applications {
    application = "myapp"
    privileges  = ["admin", "read"]
    resources   = ["*"]
  }

  run_as = ["other_user"]

  metadata = jsonencode({
    version = 1
  })
}

data "elasticstack_elasticsearch_security_role" "test" {
  name = elasticstack_elasticsearch_security_role.test.name
}
`
