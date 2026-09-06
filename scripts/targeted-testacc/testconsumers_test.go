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
	"reflect"
	"testing"
)

func TestFindTestConsumersMulti_SyntheticTree(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/kibana/space/resource.go", "package space")
	writeFile(t, root, "internal/kibana/space/resource_test.go", "package space\n\n// uses elasticstack_kibana_space\n")
	writeFile(t, root, "internal/kibana/dashboard/dashboard_test.go", "package dashboard\n\n// uses elasticstack_kibana_dashboard\n")
	writeFile(t, root, "internal/fleet/policy/resource.go", "package policy")
	writeFile(t, root, "internal/fleet/policy/testdata/main.tf", "# elasticstack_fleet_agent_policy\n")
	writeFile(t, root, "internal/other/helper.go", "package other")

	names := []string{
		"elasticstack_kibana_space",
		"elasticstack_kibana_dashboard",
		"elasticstack_fleet_agent_policy",
	}

	got, err := FindTestConsumersMulti("internal", "github.com/example/mod", names)
	if err != nil {
		t.Fatalf("FindTestConsumersMulti: %v", err)
	}

	want := map[string][]string{
		"github.com/example/mod/internal/fleet/policy":     {"elasticstack_fleet_agent_policy"},
		"github.com/example/mod/internal/kibana/dashboard": {"elasticstack_kibana_dashboard"},
		"github.com/example/mod/internal/kibana/space":     {"elasticstack_kibana_space"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("consumers = %v, want %v", got, want)
	}
}

func TestFindTestConsumersMulti_ReportsOnlyMatchedNames(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	// The package mentions only elasticstack_kibana_space even though both
	// space and spaces were candidates; the report must list only the name
	// that actually matched.
	writeFile(t, root, "internal/fleet/agentpolicy/resource.go", "package agentpolicy")
	writeFile(t, root, "internal/fleet/agentpolicy/testdata/main.tf", "# elasticstack_kibana_space\n")

	got, err := FindTestConsumersMulti("internal", "github.com/example/mod", []string{"elasticstack_kibana_space", "elasticstack_kibana_spaces"})
	if err != nil {
		t.Fatalf("FindTestConsumersMulti: %v", err)
	}
	want := map[string][]string{
		"github.com/example/mod/internal/fleet/agentpolicy": {"elasticstack_kibana_space"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("consumers = %v, want %v", got, want)
	}
}

func TestFindTestConsumersMulti_NoMatches(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	writeFile(t, root, "internal/a/a.go", "package a")
	writeFile(t, root, "internal/a/a_test.go", "package a\n\nfunc TestA(t *testing.T) {}\n")

	got, err := FindTestConsumersMulti("internal", "github.com/example/mod", []string{"elasticstack_kibana_nope"})
	if err != nil {
		t.Fatalf("FindTestConsumersMulti: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("consumers = %v, want empty", got)
	}
}
