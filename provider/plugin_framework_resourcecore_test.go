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

package provider

import (
	"context"
	"reflect"
	"testing"

	apikey "github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security/api_key"
	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPluginFrameworkResourcesEmbedResourceCore(t *testing.T) {
	t.Parallel()

	p := &Provider{version: "test"}
	ctx := context.Background()

	factories := make([]func() resource.Resource, 0, len(p.resources(ctx))+len(p.experimentalResources(ctx)))
	factories = append(factories, p.resources(ctx)...)
	factories = append(factories, p.experimentalResources(ctx)...)

	corePtrType := reflect.TypeFor[*resourcecore.Core]()

	for i, newRes := range factories {
		r := newRes()
		if _, ok := r.(*apikey.Resource); ok {
			// Out of scope for the resourcecore rollout: Configure mutates package state.
			continue
		}
		rt := reflect.TypeOf(r)
		if !typeEmbedsCorePtr(rt, corePtrType) {
			t.Fatalf("resource %d (%s) does not embed *resourcecore.Core", i, rt.String())
		}
	}
}

func typeEmbedsCorePtr(rt reflect.Type, corePtr reflect.Type) bool {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return false
	}
	for f := range rt.Fields() {
		f := f
		if f.Anonymous && f.Type == corePtr {
			return true
		}
	}
	return false
}
