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
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("generate-skill", flag.ContinueOnError)
	fs.SetOutput(stderr)

	docsDir := fs.String("docs", "docs", "Path to docs directory (tfplugindocs output)")
	assetsDir := fs.String("assets", "scripts/generate-skill/assets", "Path to hand-seeded static content")
	outDir := fs.String("out", "dist/skill/elasticstack-terraform", "Output directory for the generated skill")
	providerVersion := fs.String("provider-version", "", "Provider version to stamp into metadata.version (e.g. 0.14.3). Falls back to 0.0.0-dev when empty.")
	verbose := fs.Bool("v", false, "Verbose logging")

	if err := fs.Parse(args); err != nil {
		return err
	}

	entities, err := loadEntities(*docsDir)
	if err != nil {
		return fmt.Errorf("load entities: %w", err)
	}
	if *verbose {
		fmt.Fprintf(stdout, "loaded %d entities from docs/\n", len(entities))
	}

	gen := &generator{
		entities:        entities,
		docsDir:         *docsDir,
		assetsDir:       *assetsDir,
		outDir:          *outDir,
		providerVersion: *providerVersion,
		log:             stdout,
		verbose:         *verbose,
	}
	if err := gen.emit(); err != nil {
		return fmt.Errorf("emit skill: %w", err)
	}
	fmt.Fprintf(stdout, "wrote skill to %s (%d entities)\n", *outDir, len(entities))
	return nil
}
