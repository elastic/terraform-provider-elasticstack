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
	"strings"
	"testing"
)

func TestRun_missingSubcommand(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	err := run([]string{}, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	out := stderr.String()
	if !strings.Contains(out, "Usage: changelog") {
		t.Fatalf("expected usage output, stderr:\n%s", out)
	}
}

func TestRun_unknownSubcommand(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	err := run([]string{"not-a-real-subcommand"}, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
	out := stderr.String()
	if !strings.Contains(out, "Usage: changelog") {
		t.Fatalf("expected usage output, stderr:\n%s", out)
	}
	if !strings.Contains(out, "gather-evidence") {
		t.Fatalf("expected listed subcommands on stderr:\n%s", out)
	}
}

func TestRun_knownStub(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	err := run([]string{"run-engine"}, &stderr)
	if err == nil || !strings.Contains(err.Error(), "not yet implemented") {
		t.Fatalf("expected not-yet-implemented error, got %v", err)
	}
}
