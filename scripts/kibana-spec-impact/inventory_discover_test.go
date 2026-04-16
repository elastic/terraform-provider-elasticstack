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

const expectedSDKKibanaPkg = "github.com/elastic/terraform-provider-elasticstack/internal/kibana"

func TestSdkKibanaPkgPath(t *testing.T) {
	// Table: known SDK entities and a synthetic name must all map to root internal/kibana.
	for _, name := range []string{
		"elasticstack_kibana_space",
		"elasticstack_kibana_security_role",
		"elasticstack_kibana_action_connector",
		"elasticstack_kibana_hypothetical_sdk_only",
	} {
		t.Run(name, func(t *testing.T) {
			if got := sdkKibanaPkgPath(name); got != expectedSDKKibanaPkg {
				t.Fatalf("got %q want %q", got, expectedSDKKibanaPkg)
			}
		})
	}
}

func TestDiscoverKibanaEntitiesSmoke(t *testing.T) {
	entities := discoverKibanaEntities()
	if len(entities) < 5 {
		t.Fatalf("expected multiple Kibana entities from provider registration, got %d", len(entities))
	}
	seen := make(map[string]struct{}, len(entities))
	for _, e := range entities {
		if !strings.HasPrefix(e.Name, kibanaEntityPrefix) {
			t.Errorf("entity %q: expected %q prefix", e.Name, kibanaEntityPrefix)
		}
		if e.Type != "resource" && e.Type != "data source" {
			t.Errorf("entity %q: unexpected type %q", e.Name, e.Type)
		}
		if e.PkgPath == "" {
			t.Errorf("entity %q: empty pkg_path", e.Name)
		}
		seen[e.Name] = struct{}{}
	}
	if _, ok := seen["elasticstack_kibana_space"]; !ok {
		t.Fatal("expected stable entity elasticstack_kibana_space in inventory")
	}
}
