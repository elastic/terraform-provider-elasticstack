// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package engine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/section"
)

// Mode identifiers for changelog engine runs.
const (
	ModeUnreleased = "unreleased"
	ModeRelease    = "release"

	branchGeneratedChangelog = "generated-changelog"
	headerUnreleasedMarkdown = "## [Unreleased]"
)

var targetSemverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidateModeAndTargetVersion mirrors validateModeAndTargetVersion in changelog-engine-factory.js.
func ValidateModeAndTargetVersion(mode, targetVersion string) error {
	if mode != ModeUnreleased && mode != ModeRelease {
		return fmt.Errorf(`invalid changelog mode: %q (must be 'unreleased' or 'release')`, mode)
	}

	if mode == ModeRelease &&
		(targetVersion == "" || !targetSemverPattern.MatchString(targetVersion)) {
		return fmt.Errorf(
			"release mode requires a non-empty semver target version X.Y.Z without a leading 'v'",
		)
	}

	return nil
}

// FormatAssemblyFailureMessage mirrors formatAssemblyFailureMessage in changelog-engine-factory.js.
func FormatAssemblyFailureMessage(errors []section.AssemblyError) string {
	var msgs []string
	for _, e := range errors {
		msgs = append(msgs, fmt.Sprintf(`  - %s`, e.Reason))
	}
	var errorMessages strings.Builder
	for i, m := range msgs {
		if i > 0 {
			errorMessages.WriteString("\n")
		}
		errorMessages.WriteString(m)
	}
	return fmt.Sprintf(
		"Changelog assembly failed. The following pull requests are missing a "+
			"required ## Changelog section or Summary field:\n%s\n\n"+
			"Each merged PR must either:\n"+
			`  1. Have a '## Changelog' section with 'Customer impact' and (when not 'none') a 'Summary' field, OR`+"\n"+
			`  2. Be labeled 'no-changelog'`,
		errorMessages.String(),
	)
}
