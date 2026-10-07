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
	"strings"
)

const stackVersionKey = "STACK_VERSION="

// LatestGA returns the highest non-SNAPSHOT version in the list.
func LatestGA(versions []string) (string, error) {
	latest := ""
	for _, v := range versions {
		if strings.HasSuffix(v, "-SNAPSHOT") {
			continue
		}
		if latest == "" || versionLess(latest, v) {
			latest = v
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no GA version found in matrix")
	}
	return latest, nil
}

// SetStackVersion rewrites the STACK_VERSION= line in env-file content and
// reports whether the content changed. All other lines are preserved.
func SetStackVersion(content, version string) (string, bool, error) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, stackVersionKey) {
			continue
		}
		want := stackVersionKey + version
		if line == want {
			return content, false, nil
		}
		lines[i] = want
		return strings.Join(lines, "\n"), true, nil
	}
	return "", false, fmt.Errorf("no %s line found", stackVersionKey)
}

// SyncEnvFile sets STACK_VERSION in envPath to the latest GA version in the
// pinned artifact. It returns the version and whether the file was modified.
func SyncEnvFile(artifactPath, envPath string) (string, bool, error) {
	versions, err := ReadArtifact(artifactPath)
	if err != nil {
		return "", false, err
	}
	latest, err := LatestGA(versions)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		return "", false, fmt.Errorf("read env file: %w", err)
	}
	updated, changed, err := SetStackVersion(string(data), latest)
	if err != nil {
		return "", false, fmt.Errorf("%s: %w", envPath, err)
	}
	if !changed {
		return latest, false, nil
	}
	if err := os.WriteFile(envPath, []byte(updated), 0o644); err != nil {
		return "", false, fmt.Errorf("write env file: %w", err)
	}
	return latest, true, nil
}
