package acctest

import (
	"path"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Providers is a stub for the real acceptance test provider factory map.
var Providers = map[string]func() (resource.ProviderServer, error){}

// NamedTestCaseDirectory returns a TestStepConfigFunc that resolves to a named subdirectory.
func NamedTestCaseDirectory(name string) config.TestStepConfigFunc {
	return func(tscr config.TestStepConfigRequest) string {
		return path.Join(config.TestNameDirectory()(tscr), name)
	}
}
