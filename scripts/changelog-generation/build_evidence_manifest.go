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

package main

import (
	"fmt"
	"os"
)

// runBuildEvidenceManifest is a stub. The real implementation is in the inline JS
// gather-pr-evidence step in .github/workflows-src/changelog-generation/scripts/.
func runBuildEvidenceManifest(_ []string) {
	fmt.Fprintln(os.Stderr, "build-evidence-manifest: this subcommand is a stub.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "The real implementation runs as an inline GitHub Actions step using:")
	fmt.Fprintln(os.Stderr, "  .github/workflows-src/changelog-generation/scripts/gather-pr-evidence.inline.js")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  GITHUB_TOKEN=...  \\")
	fmt.Fprintln(os.Stderr, "  GITHUB_REPOSITORY=owner/repo  \\")
	fmt.Fprintln(os.Stderr, "  node .github/workflows-src/changelog-generation/scripts/gather-pr-evidence.inline.js")
	os.Exit(1)
}
