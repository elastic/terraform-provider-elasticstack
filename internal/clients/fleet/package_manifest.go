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

package fleet

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	packageNameRe    = regexp.MustCompile(`(?m)^name:\s*(\S+)`)
	packageVersionRe = regexp.MustCompile(`(?m)^version:\s*["']?([^\s"']+)["']?`)
)

// parsePackageInfo parses the package name and version from the manifest.yml
// inside a package archive. It dispatches to the appropriate parser based on
// the file extension (.zip or .tar.gz / .gz).
func parsePackageInfo(path string) (name, version string, err error) {
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		return parsePackageInfoFromTarGz(path)
	}
	return parsePackageInfoFromZip(path)
}

// isManifestYAML reports whether entryName is a top-level or direct-child manifest.yml.
func isManifestYAML(entryName string) bool {
	return strings.HasSuffix(entryName, "/manifest.yml") || entryName == "manifest.yml"
}

// extractPackageNameVersion parses the name and version fields from manifest YAML content.
func extractPackageNameVersion(content []byte) (name, version string) {
	if m := packageNameRe.FindSubmatch(content); len(m) >= 2 {
		name = string(m[1])
	}
	if m := packageVersionRe.FindSubmatch(content); len(m) >= 2 {
		version = string(m[1])
	}
	return name, version
}

// parsePackageInfoFromZip opens a zip archive at path, finds the top-level
// manifest.yml, and extracts the package name and version fields. It is used as
// a fallback when the Fleet upload API response does not include the package
// name or version (older Kibana versions).
func parsePackageInfoFromZip(path string) (name, version string, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", "", fmt.Errorf("opening zip %q: %w", path, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !isManifestYAML(f.Name) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", "", fmt.Errorf("opening manifest.yml in zip: %w", err)
		}
		content, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			return "", "", fmt.Errorf("reading manifest.yml: %w", readErr)
		}
		name, version = extractPackageNameVersion(content)
		if name != "" {
			return name, version, nil
		}
	}
	return "", "", fmt.Errorf("manifest.yml with name field not found in zip")
}

// parsePackageInfoFromTarGz opens a gzip-compressed tar archive at path, finds
// the top-level manifest.yml, and extracts the package name and version fields.
// It is used as a fallback for tar.gz archives when the Fleet upload API
// response does not include the package name or version (older Kibana versions).
func parsePackageInfoFromTarGz(path string) (name, version string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", fmt.Errorf("opening tar.gz %q: %w", path, err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", "", fmt.Errorf("creating gzip reader for %q: %w", path, err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("reading tar.gz %q: %w", path, err)
		}
		if !isManifestYAML(hdr.Name) {
			continue
		}
		content, readErr := io.ReadAll(tr)
		if readErr != nil {
			return "", "", fmt.Errorf("reading manifest.yml from tar.gz: %w", readErr)
		}
		name, version = extractPackageNameVersion(content)
		if name != "" {
			return name, version, nil
		}
	}
	return "", "", fmt.Errorf("manifest.yml with name field not found in tar.gz")
}
