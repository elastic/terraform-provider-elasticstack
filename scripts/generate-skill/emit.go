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
	"io"
	"os"
	"path/filepath"
	"strings"
)

type generator struct {
	entities        []*entity
	docsDir         string
	assetsDir       string
	outDir          string
	providerVersion string
	log             io.Writer
	verbose         bool
}

func (g *generator) emit() error {
	if err := os.RemoveAll(g.outDir); err != nil {
		return fmt.Errorf("clean out dir: %w", err)
	}
	dirs := []string{
		g.outDir,
		filepath.Join(g.outDir, "references"),
		filepath.Join(g.outDir, "references", "resources"),
		filepath.Join(g.outDir, "references", "data-sources"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}

	if err := g.copyStaticAssets(); err != nil {
		return err
	}
	if err := g.writeIndex(); err != nil {
		return err
	}
	if err := g.copyDocFiles(); err != nil {
		return err
	}
	if err := g.writeProvenance(); err != nil {
		return err
	}
	return nil
}

const indexEntitiesPlaceholder = "{{ENTITIES}}"

func (g *generator) writeIndex() error {
	path := filepath.Join(g.outDir, "references", "index.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read index template: %w", err)
	}
	if !strings.Contains(string(raw), indexEntitiesPlaceholder) {
		return fmt.Errorf("references/index.md template is missing %s placeholder", indexEntitiesPlaceholder)
	}
	rendered := strings.Replace(string(raw), indexEntitiesPlaceholder, g.renderEntityList(), 1)
	return os.WriteFile(path, []byte(rendered), 0o644)
}

func (g *generator) renderEntityList() string {
	var b strings.Builder
	for _, e := range g.entities {
		if e.Kinds.has(kindResource) {
			fmt.Fprintf(&b, "- `%s` (resource) — %s → references/resources/%s.md\n", e.Name, oneLine(e.ResourceSummary), e.ShortName)
		}
		if e.Kinds.has(kindDataSource) {
			fmt.Fprintf(&b, "- `%s` (data source) — %s → references/data-sources/%s.md\n", e.Name, oneLine(e.DataSrcSummary), e.ShortName)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// copyDocFiles copies docs/resources/*.md and docs/data-sources/*.md verbatim
// into references/resources/ and references/data-sources/.
func (g *generator) copyDocFiles() error {
	for _, sub := range []string{"resources", "data-sources"} {
		src := filepath.Join(g.docsDir, sub)
		dst := filepath.Join(g.outDir, "references", sub)
		des, err := os.ReadDir(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read %s: %w", src, err)
		}
		for _, de := range des {
			if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(src, de.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dst, de.Name()), data, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) copyStaticAssets() error {
	if _, err := os.Stat(g.assetsDir); os.IsNotExist(err) {
		fmt.Fprintf(g.log, "note: assets dir %q does not exist; skipping static content\n", g.assetsDir)
		return nil
	}
	return filepath.WalkDir(g.assetsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(g.assetsDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(g.outDir, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".md") {
			data = []byte(g.applySubstitutions(string(data)))
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

func (g *generator) applySubstitutions(s string) string {
	version := g.providerVersion
	if version == "" {
		version = "0.0.0-dev"
	}
	return strings.ReplaceAll(s, "{{VERSION}}", version)
}

func (g *generator) writeProvenance() error {
	var b strings.Builder
	b.WriteString("# Generated skill — provenance\n\n")
	b.WriteString("This skill is auto-generated from the elastic/terraform-provider-elasticstack repository.\n\n")
	fmt.Fprintf(&b, "- Entities: %d (resources: %d, data sources: %d)\n",
		len(g.entities), countKind(g.entities, kindResource), countKind(g.entities, kindDataSource))
	if g.providerVersion != "" {
		b.WriteString("- Provider version: " + g.providerVersion + "\n")
	}
	b.WriteString("- Source of truth: `docs/`\n")
	b.WriteString("- Generator: `scripts/generate-skill`\n\n")
	b.WriteString("Do not edit files in this directory by hand. Re-run `make skill-generate` from the provider repo instead.\n")
	return writeFile(filepath.Join(g.outDir, "GENERATED.md"), b.String())
}

// --- helpers ---

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func countKind(es []*entity, k entityKind) int {
	n := 0
	for _, e := range es {
		if e.Kinds.has(k) {
			n++
		}
	}
	return n
}

func oneLine(s string) string {
	var line string
	for l := range strings.SplitSeq(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			line = l
			break
		}
	}
	if line == "" {
		return ""
	}
	if i := strings.Index(line, ". "); i > 0 {
		line = line[:i+1]
	}
	for _, marker := range []string{" See:", " see:", " See the", " see the"} {
		if i := strings.Index(line, marker); i > 0 {
			line = line[:i]
		}
	}
	return strings.TrimSpace(line)
}

