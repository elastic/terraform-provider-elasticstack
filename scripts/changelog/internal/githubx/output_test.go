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

package githubx_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
)

func TestAppendGitHubOutput_singleLine(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out")
	if err := githubx.AppendGitHubOutput(path, "status", "ok"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(raw), "status=ok\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestAppendGitHubOutput_multilineHeredoc(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "out")
	content := `{
  "a": 1,
  "b": 2
}`
	if err := githubx.AppendGitHubOutput(path, "result_json", content); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(raw)
	if !strings.HasPrefix(got, "result_json<<") {
		t.Fatalf("expected heredoc start, got %q", got)
	}
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("unexpected heredoc shape: %q", got)
	}
	re := regexp.MustCompile(`^result_json<<([a-f0-9]+)$`)
	sm := re.FindStringSubmatch(lines[0])
	if sm == nil {
		t.Fatalf("first line should be result_json<<<delimiter>, got %q", lines[0])
	}
	delim := sm[1]
	if lines[len(lines)-1] != delim {
		t.Fatalf("last line must repeat delimiter %q: %v", delim, lines)
	}
	body := strings.Join(lines[1:len(lines)-1], "\n")
	if body != content {
		t.Fatalf("body mismatch\ngot:\n%s\nwant:\n%s", body, content)
	}
}

func TestAppendGitHubOutput_errors(t *testing.T) {
	t.Parallel()
	if err := githubx.AppendGitHubOutput("", "k", "v"); err == nil {
		t.Fatal("expected error for empty path")
	}
	if err := githubx.AppendGitHubOutput(t.TempDir()+"/x", "", "v"); err == nil {
		t.Fatal("expected error for empty name")
	}
}
