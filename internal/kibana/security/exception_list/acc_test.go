package exception_list_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minExceptionListAPISupport = version.Must(version.NewVersion("7.9.0"))

func TestAccResourceExceptionList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				Config:   testAccResourceExceptionListCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "list_id", "test-exception-list"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Test exception list for acceptance tests"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "type", "detection"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "namespace_type", "single"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.0", "test"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "id"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_at"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_security_exception_list.test", "created_by"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minExceptionListAPISupport),
				Config:   testAccResourceExceptionListUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "name", "Test Exception List Updated"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "description", "Updated description"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.0", "test"),
					resource.TestCheckResourceAttr("elasticstack_kibana_security_exception_list.test", "tags.1", "updated"),
				),
			},
		},
	})
}

const testAccResourceExceptionListCreate = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = "test-exception-list"
  name           = "Test Exception List"
  description    = "Test exception list for acceptance tests"
  type           = "detection"
  namespace_type = "single"
  
  tags = ["test"]
}
`

const testAccResourceExceptionListUpdate = `
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_exception_list" "test" {
  list_id        = "test-exception-list"
  name           = "Test Exception List Updated"
  description    = "Updated description"
  type           = "detection"
  namespace_type = "single"
  
  tags = ["test", "updated"]
}
`
