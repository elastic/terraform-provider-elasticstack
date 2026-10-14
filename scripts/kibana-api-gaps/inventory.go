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
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const entitycorePkgSuffix = "/internal/entitycore"

// Resource describes a Terraform resource registered on the provider.
type Resource struct {
	Name    string `json:"name"`
	PkgPath string `json:"pkg_path"`
}

// discoverResources lists the resources registered on the provider (data sources are not
// included). Whether a resource calls the Kibana API is decided later by reachability analysis.
func discoverResources() []Resource {
	ctx := context.Background()
	var out []Resource
	for _, factory := range provider.NewFrameworkProvider("kibana-api-gaps").Resources(ctx) {
		r := factory()
		var meta resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &meta)
		if meta.TypeName == "" {
			continue
		}
		out = append(out, Resource{Name: meta.TypeName, PkgPath: owningPackage(r, factory)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// owningPackage returns the package defining the resource. Resources built directly from the
// generic entitycore envelope are attributed to the package of their factory function instead.
func owningPackage(r resource.Resource, factory func() resource.Resource) string {
	t := reflect.TypeOf(r)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if pkg := t.PkgPath(); pkg != "" && !strings.HasSuffix(pkg, entitycorePkgSuffix) {
		return pkg
	}
	return factoryPackage(runtime.FuncForPC(reflect.ValueOf(factory).Pointer()).Name())
}

func factoryPackage(funcName string) string {
	slash := strings.LastIndex(funcName, "/")
	dot := strings.Index(funcName[slash+1:], ".")
	if dot < 0 {
		return funcName
	}
	return funcName[:slash+1+dot]
}
