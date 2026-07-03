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
	"reflect"
	"testing"
)

func TestFindAccTestPackages_SyntheticTree(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/kibana/space/resource_test.go", "package space\n\nfunc TestAccSpace_basic(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/space/resource_unit_test.go", "package space\n\nfunc TestUnitSpace(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/dashboard/dashboard_test.go", "package dashboard\n\nfunc TestAccDashboard(t *testing.T) {}\n")
	writeFile(t, root, "internal/kibana/dashboard/panel/lens/lens_test.go", "package lens\n\nfunc TestAccLensPanel(t *testing.T) {}\n")
	writeFile(t, root, "internal/fleet/policy/resource.go", "package policy\n")
	writeFile(t, root, "internal/fleet/policy/resource_test.go", "package policy\n\nfunc TestAccPolicy(t *testing.T) {}\n")
	// No _test.go file at all.
	writeFile(t, root, "internal/pkg/resource.go", "package pkg\n")

	got, err := FindAccTestPackages("internal", "github.com/example/mod")
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}

	want := []string{
		"github.com/example/mod/internal/fleet/policy",
		"github.com/example/mod/internal/kibana/dashboard",
		"github.com/example/mod/internal/kibana/dashboard/panel/lens",
		"github.com/example/mod/internal/kibana/space",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FindAccTestPackages = %v, want %v", got, want)
	}
}

func TestFindAccTestPackages_NoPackages(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "foo.go", "package foo\n")

	got, err := FindAccTestPackages(".", "github.com/example/mod")
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no packages, got %v", got)
	}
}

func TestIsAccTestFile(t *testing.T) {
	root := t.TempDir()

	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "with-acc-test",
			content: "package foo\n\nfunc TestAccSomething(t *testing.T) {}\n",
			want:    true,
		},
		{
			name:    "unit-test-only",
			content: "package foo\n\nfunc TestSomething(t *testing.T) {}\n",
			want:    false,
		},
		{
			name:    "lowercase-acc",
			content: "package foo\n\nfunc TestaccSomething(t *testing.T) {}\n",
			want:    false,
		},
		{
			name:    "acc-in-comment",
			content: "package foo\n\n// func TestAccSomething(t *testing.T) {}\n",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, tc.name+".go")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatalf("write file: %v", err)
			}
			got, err := isAccTestFile(path)
			if err != nil {
				t.Fatalf("isAccTestFile: %v", err)
			}
			if got != tc.want {
				t.Errorf("isAccTestFile = %v, want %v", got, tc.want)
			}
		})
	}
}
