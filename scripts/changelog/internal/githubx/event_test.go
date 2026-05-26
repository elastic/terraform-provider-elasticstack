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
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/githubx"
)

func TestReadEventPayload_emptyPath(t *testing.T) {
	t.Parallel()
	_, err := githubx.ReadEventPayload("")
	if err == nil || !strings.Contains(err.Error(), "GITHUB_EVENT_PATH") {
		t.Fatalf("expected GITHUB_EVENT_PATH error, got %v", err)
	}
}

func TestReadEventPayload_validJSONFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "event.json")
	want := []byte(`{"action":"opened","number":1}`)
	if err := os.WriteFile(path, want, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := githubx.ReadEventPayload(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoadEvent_malformedJSONFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "event.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := githubx.LoadEvent[map[string]any](path)
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
}

func TestOptionalPullRequestNumberFromEventPath_emptyPath(t *testing.T) {
	t.Parallel()
	n, err := githubx.OptionalPullRequestNumberFromEventPath("")
	if err != nil || n != 0 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}

func TestOptionalPullRequestNumberFromEventPath_emptyObject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "event.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := githubx.OptionalPullRequestNumberFromEventPath(path)
	if err != nil || n != 0 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}

func TestOptionalPullRequestNumberFromEventPath_withPullNumber(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "event.json")
	payload := []byte(`{"pull_request":{"number":42}}`)
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := githubx.OptionalPullRequestNumberFromEventPath(path)
	if err != nil || n != 42 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}

func TestOptionalPullRequestNumberFromEventPath_invalidJSON_file(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "event.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := githubx.OptionalPullRequestNumberFromEventPath(path)
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal-related error, got %v", err)
	}
}

func TestDecodeEvent_malformedJSON(t *testing.T) {
	t.Parallel()
	_, err := githubx.DecodeEvent[map[string]any]([]byte(`{`))
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatalf("expected unmarshal error, got %v", err)
	}
}
