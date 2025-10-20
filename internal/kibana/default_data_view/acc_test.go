package default_data_view_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))

func TestAccResourceDefaultDataView(t *testing.T) {
	indexName1 := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName2 := "my-other-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   testAccResourceDefaultDataViewBasic(indexName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   testAccResourceDefaultDataViewUpdate(indexName1, indexName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithSkipDelete(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   testAccResourceDefaultDataViewWithSkipDelete(indexName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
		},
	})
}

func TestAccResourceDefaultDataViewWithCustomSpace(t *testing.T) {
	indexName := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	spaceID := "test-space-" + sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   testAccResourceDefaultDataViewWithCustomSpace(indexName, spaceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", spaceID),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "force", "true"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "skip_delete", "false"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", spaceID),
				),
			},
		},
	})
}

func testAccResourceDefaultDataViewBasic(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	data_view = {
		title = "%s*"
	}
	depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
	data_view_id = elasticstack_kibana_data_view.dv.data_view.id
	force        = true
}
`, indexName, indexName)
}

func testAccResourceDefaultDataViewUpdate(indexName1, indexName2 string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_elasticsearch_index" "my_other_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	data_view = {
		title = "%s*"
	}
	depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_data_view" "dv2" {
	data_view = {
		title = "%s*"
	}
	depends_on = [elasticstack_elasticsearch_index.my_other_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
	data_view_id = elasticstack_kibana_data_view.dv2.data_view.id
	force        = true
}
`, indexName1, indexName2, indexName1, indexName2)
}

func testAccResourceDefaultDataViewWithSkipDelete(indexName string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	data_view = {
		title = "%s*"
	}
	depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
	data_view_id = elasticstack_kibana_data_view.dv.data_view.id
	force        = true
	skip_delete  = true
}
`, indexName, indexName)
}

func testAccResourceDefaultDataViewWithCustomSpace(indexName, spaceID string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
	elasticsearch {}
	kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
	space_id    = "%s"
	name        = "Test Space %s"
	description = "Test space for default data view"
}

resource "elasticstack_elasticsearch_index" "my_index" {
	name                = "%s"
	deletion_protection = false
}

resource "elasticstack_kibana_data_view" "dv" {
	space_id = elasticstack_kibana_space.test_space.space_id
	data_view = {
		title = "%s*"
	}
	depends_on = [elasticstack_elasticsearch_index.my_index]
}

resource "elasticstack_kibana_default_data_view" "test" {
	space_id     = elasticstack_kibana_space.test_space.space_id
	data_view_id = elasticstack_kibana_data_view.dv.data_view.id
	force        = true
}
`, spaceID, spaceID, indexName, indexName)
}
