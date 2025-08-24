package ingest_test

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// check if the provided json string equal to the generated one
func CheckResourceJson(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		v, ok := is.Attributes[key]
		if !ok {
			return fmt.Errorf("%s: Attribute '%s' not found", name, key)
		}
		if eq, err := utils.JSONBytesEqual([]byte(value), []byte(v)); !eq {
			return fmt.Errorf(
				"%s: Attribute '%s' expected %#v, got %#v (<err>: %v)",
				name,
				key,
				value,
				v,
				err)
		}
		return nil
	}
}
