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
	"context"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const kibanaEntityPrefix = "elasticstack_kibana_"

// Entity describes a Terraform Kibana entity derived from provider registrations.
type Entity struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	PkgPath string `json:"pkg_path"`
}

// discoverKibanaEntities lists resources and data sources registered on the provider whose
// Terraform type name uses the elasticstack_kibana_ prefix (Plugin Framework + Plugin SDK).
func discoverKibanaEntities() []Entity {
	fwProv := provider.NewFrameworkProvider("kibana-spec-impact")
	sdkProv := provider.New("kibana-spec-impact")
	ctx := context.Background()

	var out []Entity
	seen := make(map[string]struct{})

	for _, rf := range fwProv.Resources(ctx) {
		r := rf()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
		if meta.TypeName == "" || !strings.HasPrefix(meta.TypeName, kibanaEntityPrefix) {
			continue
		}
		pkg := reflect.TypeOf(r).Elem().PkgPath()
		out = append(out, Entity{Type: "resource", Name: meta.TypeName, PkgPath: pkg})
		seen[meta.TypeName] = struct{}{}
	}

	for _, dsf := range fwProv.DataSources(ctx) {
		ds := dsf()
		var meta datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
		if meta.TypeName == "" || !strings.HasPrefix(meta.TypeName, kibanaEntityPrefix) {
			continue
		}
		pkg := reflect.TypeOf(ds).Elem().PkgPath()
		out = append(out, Entity{Type: "data source", Name: meta.TypeName, PkgPath: pkg})
		seen[meta.TypeName] = struct{}{}
	}

	for name := range sdkProv.ResourcesMap {
		if !strings.HasPrefix(name, kibanaEntityPrefix) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		pkg := sdkKibanaPkgPath(name)
		out = append(out, Entity{Type: "resource", Name: name, PkgPath: pkg})
		seen[name] = struct{}{}
	}

	for name := range sdkProv.DataSourcesMap {
		if !strings.HasPrefix(name, kibanaEntityPrefix) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		pkg := sdkKibanaPkgPath(name)
		out = append(out, Entity{Type: "data source", Name: name, PkgPath: pkg})
		seen[name] = struct{}{}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// sdkKibanaPkgPath returns the Go import path for SDK-only Kibana entities in the root
// internal/kibana package. New SDK Kibana entities use the same default without failing inventory.
func sdkKibanaPkgPath(_ string) string {
	return "github.com/elastic/terraform-provider-elasticstack/internal/kibana"
}

// entityScanPaths returns Go files or directories (under repoRoot) to scan for kbapi/kibanaoapi usage.
func entityScanPaths(repoRoot string, e Entity) ([]string, error) {
	root := filepath.Clean(repoRoot)
	if !strings.Contains(e.PkgPath, "/internal/kibana") {
		return nil, nil
	}
	rel := strings.TrimPrefix(e.PkgPath, "github.com/elastic/terraform-provider-elasticstack/")
	dir := filepath.Join(root, rel)
	if strings.HasSuffix(e.PkgPath, "/internal/kibana") {
		prefixes := rootKibanaFilePrefixes(e.Name)
		if len(prefixes) == 0 {
			// Unknown new entity under root internal/kibana: scan the full package directory
			// rather than failing on prefix mapping.
			return []string{dir}, nil
		}
		var paths []string
		for _, p := range prefixes {
			matches, err := filepath.Glob(filepath.Join(dir, p+"*.go"))
			if err != nil {
				return nil, err
			}
			paths = append(paths, matches...)
		}
		if len(paths) == 0 {
			return []string{dir}, nil
		}
		return paths, nil
	}
	return []string{dir}, nil
}

// rootKibanaFilePrefixes returns filename prefixes for narrow scans within root internal/kibana.
// An empty slice means the caller should scan the whole package directory.
func rootKibanaFilePrefixes(entityName string) []string {
	s := strings.TrimPrefix(entityName, kibanaEntityPrefix)
	switch s {
	case "space":
		return []string{"space"}
	case "security_role":
		return []string{"role"}
	case "action_connector":
		return []string{"connector"}
	default:
		return nil
	}
}
