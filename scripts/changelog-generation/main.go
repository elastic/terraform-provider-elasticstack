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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate-provenance":
		runValidateProvenance(os.Args[2:])
	case "rewrite-changelog-section":
		runRewriteChangelogSection(os.Args[2:])
	case "build-evidence-manifest":
		runBuildEvidenceManifest(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %q\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./scripts/changelog-generation <subcommand> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  validate-provenance       Validate provenance JSON against evidence manifest")
	fmt.Fprintln(os.Stderr, "  rewrite-changelog-section Rewrite a section in CHANGELOG.md")
	fmt.Fprintln(os.Stderr, "  build-evidence-manifest   Build evidence manifest (stub; real impl is inline JS)")
}
