package acctest

import (
	"path"

	"github.com/hashicorp/terraform-plugin-testing/config"
)

// NamedTestCaseDirectory returns a TestStepConfigFunc that resolves to a named subdirectory.
func NamedTestCaseDirectory(name string) config.TestStepConfigFunc {
	return func(tscr config.TestStepConfigRequest) string {
		return path.Join(config.TestNameDirectory()(tscr), name)
	}
}
