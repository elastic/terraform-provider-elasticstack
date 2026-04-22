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
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeOK(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func stubKibanaOAPIPackage(t *testing.T, root string) {
	t.Helper()
	stub := filepath.Join(root, "internal", "clients", "kibanaoapi", "stub.go")
	writeOK(t, stub, []byte("package kibanaoapi\n"))
}

func TestMatchHighConfidenceDirectKbapiSelector(t *testing.T) {
	root := t.TempDir()
	stubKibanaOAPIPackage(t, root)
	entDir := filepath.Join(root, "internal", "kibana", "fixturematch")
	entityFile := filepath.Join(entDir, "resource.go")
	writeOK(t, entityFile, []byte(`package fixturematch

func Example() {
	var _ kbapi.WidgetType
}
`))

	oapi, err := buildKibanaOAPIIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	matched, err := matchHighConfidence([]string{entDir}, oapi, []string{"WidgetType"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(matched, "WidgetType") {
		t.Fatalf("got %v", matched)
	}
}

func TestMatchHighConfidenceViaKibanaOAPICall(t *testing.T) {
	root := t.TempDir()
	stubKibanaOAPIPackage(t, root)
	helper := filepath.Join(root, "internal", "clients", "kibanaoapi", "fixture_helper.go")
	writeOK(t, helper, []byte(`package kibanaoapi

func FixtureFromImpactHelper() {
	// references kbapi symbol used in impact test
	var _ kbapi.PanelKind
}
`))

	entDir := filepath.Join(root, "internal", "kibana", "fixturematch2")
	entityFile := filepath.Join(entDir, "call.go")
	writeOK(t, entityFile, []byte(`package fixturematch2

func Run() {
	kibanaoapi.FixtureFromImpactHelper()
}
`))

	oapi, err := buildKibanaOAPIIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	matched, err := matchHighConfidence([]string{entDir}, oapi, []string{"PanelKind"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(matched, "PanelKind") {
		t.Fatalf("got %v", matched)
	}
}
