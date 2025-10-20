package default_data_view_test

import (
	"embed"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//go:embed test_data/*.tf
var testDataFS embed.FS

var minDataViewAPISupport = version.Must(version.NewVersion("8.1.0"))

// loadTestData reads and returns the content of a test data file
func loadTestData(filename string) string {
	data, err := testDataFS.ReadFile("test_data/" + filename)
	if err != nil {
		panic("Failed to load test data file: " + filename + " - " + err.Error())
	}
	return string(data)
}

func TestAccResourceDefaultDataView(t *testing.T) {
	indexName1 := "my-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)
	indexName2 := "my-other-index-" + sdkacctest.RandStringFromCharSet(4, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   loadTestData("basic.tf"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName1),
				},
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
				Config:   loadTestData("update.tf"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_default_data_view.test", "data_view_id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "space_id", "default"),
				),
			},
			{
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(minDataViewAPISupport),
				Config:   loadTestData("unset.tf"),
				ConfigVariables: config.Variables{
					"index_name1": config.StringVariable(indexName1),
					"index_name2": config.StringVariable(indexName2),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_default_data_view.test", "id", "default"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_default_data_view.test", "data_view_id"),
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
				Config:   loadTestData("skip_delete.tf"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
				},
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
				Config:   loadTestData("custom_space.tf"),
				ConfigVariables: config.Variables{
					"index_name": config.StringVariable(indexName),
					"space_id":   config.StringVariable(spaceID),
				},
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
