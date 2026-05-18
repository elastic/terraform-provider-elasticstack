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

// Command changelog implements GitHub workflow helpers for changelog generation
// and PR changelog validation.
//
// Usage:
//
//	go run ./scripts/changelog <subcommand> [flags]
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "changelog: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}
	switch args[0] {
	case "gather-evidence":
		return cmdGatherEvidence(args[1:], stderr)
	case "run-engine":
		return cmdRunEngine(args[1:], stderr)
	case "manage-unreleased-pr":
		return cmdManageUnreleasedPR(args[1:], stderr)
	case "refresh-release-pr":
		return cmdRefreshReleasePR(args[1:], stderr)
	case "validate-pr-section":
		return cmdValidatePRSection(args[1:], stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: changelog <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  gather-evidence        Build evidence manifest for a release range")
	fmt.Fprintln(w, "  run-engine             Run unreleased or release changelog engine")
	fmt.Fprintln(w, "  manage-unreleased-pr   Open or update the generated-changelog PR")
	fmt.Fprintln(w, "  refresh-release-pr     Refresh the release changelog PR")
	fmt.Fprintln(w, "  validate-pr-section    Validate PR body ## Changelog section")
	return errors.New("unknown or missing subcommand")
}

func cmdGatherEvidence(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("gather-evidence", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: gather-evidence")
}

func cmdRunEngine(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("run-engine", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: run-engine")
}

func cmdManageUnreleasedPR(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("manage-unreleased-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: manage-unreleased-pr")
}

func cmdRefreshReleasePR(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("refresh-release-pr", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: refresh-release-pr")
}

func cmdValidatePRSection(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate-pr-section", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return errors.New("not yet implemented: validate-pr-section")
}
