package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestCheckResourceListAttr(name, key string, values []string) resource.TestCheckFunc {
	var testCheckFuncs []resource.TestCheckFunc
	resource.ComposeTestCheckFunc()
	for i, v := range values {
		testCheckFuncs = append(testCheckFuncs, resource.TestCheckResourceAttr(name, fmt.Sprintf("%s.%d", key, i), v))
	}
	return resource.ComposeTestCheckFunc(testCheckFuncs...)
}
