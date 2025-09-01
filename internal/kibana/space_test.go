package kibana_test

import (
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceSpace(t *testing.T) {
	spaceId := sdkacctest.RandStringFromCharSet(22, sdkacctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkResourceSpaceDestroy,
		ProtoV6ProviderFactories: acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSpaceCreate(spaceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Name %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space"),
				),
			},
			{
				Config: testAccResourceSpaceUpdate(spaceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Updated %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Updated space description"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "ingestManager"),
					resource.TestCheckTypeSetElemAttr("elasticstack_kibana_space.test_space", "disabled_features.*", "enterpriseSearch"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "color", "#FFFFFF"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_space.test_space", "image_url"),
				),
			},
			{
				Config:   testAccResourceSpaceWithSolution(spaceId),
				SkipFunc: versionutils.CheckIfVersionIsUnsupported(kibana.SpaceSolutionMinVersion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Solution %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space with Solution"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "solution", "security"),
				),
			},
			{
				Config: testAccResourceSpaceCreate(spaceId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "space_id", spaceId),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "name", fmt.Sprintf("Name %s", spaceId)),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "description", "Test Space"),
					resource.TestCheckResourceAttr("elasticstack_kibana_space.test_space", "color", "#FFFFFF"),
				),
			},
		},
	})
}

func testAccResourceSpaceCreate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "%s"
  name        = "%s"
  description = "Test Space"
}
	`, id, fmt.Sprintf("Name %s", id))
}

func testAccResourceSpaceUpdate(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id          = "%s"
  name              = "%s"
  description       = "Updated space description"
  disabled_features = ["ingestManager", "enterpriseSearch"]
  color             = "#FFFFFF"
  image_url         = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAD4AAABACAYAAABC6cT1AAAGf0lEQVRoQ+3abYydRRUH8N882xYo0IqagEVjokQJKAiKBjXExC9G/aCkGowCIghCkRcrVSSKIu/FEiqgGL6gBIlAYrAqUTH6hZgQFVEMKlQFfItWoQWhZe8z5uzMLdvbfbkLxb13d+fbvfe588x/zpn/+Z9zJpmnI81T3BaAzzfLL1h8weLzZAcWXH2eGHo7zAWLL1h8nuzAjFw9G1N6Kzq8HnuM36MR8iibF3Fv4q+7cv8yDV6K13bYq2furSP8Ag8ncr/vnSnwRViJT2GfCV7yL1yHGxLb+l3EdM9lluNEnIC9xz+f2ZL4Er6Z2DrdXN3fZwp8CU7OfDHxggle8lTLbQ1nJ/7Z7yKmey5zYGZt4h2IzR8/trRc2PDlxJPTzfVcgJ+CC0wMPOa9F6cm7up3EVM9V9386MxliVdM8GwAv6hh/awCz/w7lY25OtF5ruBz4ZLP42NYNrDAFbC3YPWuILnMAfgq3oaRQQYea/stViV+sgssvjKzLvGySeaaNVfP4d7Btokgvxj/bblgpueuF1hmWcyTCmfE3J3M1lTcv0vMswM88zR+jpw4osu6me8kzkpsfLZWzxyRuabO22buxxOJ12FxnXfWgEe83pB5sOE47BsLymzscOoi7nw2JJfZreUjiUsTyzKPZm5NvBDvSuw268AzNzV8H5/Am+qCnsAXgpgSW2Zq9cyKlksbPlTd+te4quWNieMHBfiNDdciYnwsdI/MaOaWhnMTf54J8CqNj8x8JXFIZltYu+HqlmNT8YSBsHgAPw/vxvlVV4du/s0oaxbxg0TbL/jMni0nNcVjQq7+HZfgtpbzBg342TgQ63AkmsymxBW4IjE6A+D7Vzd/fyWxIM/VuCe+HzTgZ2Jpy/kNJ2FJLmLm24mPJ/42A+Bvrxt4SISwlhsaPodH26LZB8rVA3inwwebsrixJCZzX+KMxI/7AV61eVh3DV6Mx3EOvh4kN6jAg8nfUCXm4d1wE66OyxNPTQc+s3/o/MoXizL3JE5O3F3P/uBZPPF4Zr+Wi5uSO48ZPRdyCwn7YB/A35m5KhWNHox4fcNnIs0ddOCRSBxf8+cQG+Huf0l8NJVYP+nI7NXy2ar4QqIGm69JfKPOE2w/mBavCzwM11R2D+ChsUO7hyUfmwx55qDM1xJvqZ7y08TpifuGBfjeURVJnNIVGpkNiXNS0ds7jcySDitDCCWW56LJ10fRo8sNA+3qXUSZD2CtQlZh9T+1rB7h9oliembflnMbzqgSNZKbKGHdPm7OwXb1CvQ1metSETMpszmzvikCJNh/h5E5PHNl4qga/+/cxqrdeWDYgIe7X5L4cGJPJX2940lOX8pD41FnFnc4riluvQKbK0dcHJFi2IBHNTQSlguru4d2/wPOTNzRA3x5y+U1E1uqWDkETOT026XuUJzx6u7ReLhSYenQ7uHua0fKZmwfmcPqsQjxE5WVONcRxn7X89zgn/EKPMRMxOVQXmP18Mx3q3b/Y/0cQE/IhFtHESMsHFlZ1Ml3CH3DZPHImY+pxcKumNmYirtvqMBfhMuU6s3iqOQkTsMPe1tCQwO8Ajs0lxr7W+vnp1MJc9EgCNd/cy6x+9D4veXmprj5wxMw/3C4egW6zzgZOlYZzfwo3F2J7ael0pJamvlPKgWNKFft1AAcKotXoFEbD7kaoSoQPVKB35+5KHF0lai/rJo+up87jWEE/qqqwY+qrL21LWLm95lPJ16ppKw31XC3PXYPJauPEx7B6BHCgrSizRs18qiaRp8tlN3ueCTYPHH9RNaunjI8Z7wLYpT3jZSCYXQ8e9vTsRE/q+no3XMKeObgGtaintbb/AvXj4JDkNw/5hrwYPfIvlZFUbLn7G5q+eQIN09Vnho6cqvnM/Lt99RixH49wO8K0ZL41WTWHoQzvsNVkOheZqKhEGpsp3SzB+BBtZAYve7uOR9tuTaaB6l0XScdYfEQPpkTUyHEGP+XqyDBzu+NBCITUjNWHynkrbWKOuWFn1xKzqsyx0bdvS78odp0+N503Zao0uCsWuSIDku8/7EO60b41vN5+Ses9BKlTdvd8bhp9EBvJjWJAIn/vxwHe6b3tSk6JFPV4nq85oAOrx555v/x/rh3E6Lo+bnuNS4uB4Cuq0ZfvO8X1rM6q/+vnjLVqZq7v83onttc2oYF4HPJmv1gWbB4P7s0l55ZsPhcsmY/WBYs3s8uzaVn5q3F/wf70mRuBCtbjQAAAABJRU5ErkJggg=="
}
	`, id, fmt.Sprintf("Updated %s", id))
}

func testAccResourceSpaceWithSolution(id string) string {
	return fmt.Sprintf(`
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_space" "test_space" {
  space_id    = "%s"
  name        = "%s"
  description = "Test Space with Solution"
  solution    = "security"
}
	`, id, fmt.Sprintf("Solution %s", id))
}

func checkResourceSpaceDestroy(s *terraform.State) error {
	client, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticstack_kibana_space" {
			continue
		}

		kibanaClient, err := client.GetKibanaClient()
		if err != nil {
			return err
		}
		res, err := kibanaClient.KibanaSpaces.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if res != nil {
			return fmt.Errorf("Space (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}
