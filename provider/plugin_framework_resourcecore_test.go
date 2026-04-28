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

	"github.com/elastic/terraform-provider-elasticstack/internal/resourcecore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPluginFrameworkResourcesEmbedResourceCore(t *testing.T) {
	t.Parallel()

	// Match the real registration path in (*Provider).Resources, including
	// experimental resources (same branch as acceptance tests via AccTestVersion).
	p := &Provider{version: AccTestVersion}
	ctx := context.Background()

	corePtrType := reflect.TypeFor[*resourcecore.Core]()

	for _, newRes := range p.Resources(ctx) {
		r := newRes()
		rt := reflect.TypeOf(r)
		if !typeEmbedsCorePtr(rt, corePtrType) {
			t.Fatalf("resource %T does not embed *resourcecore.Core", r)
		}
		if !resourceConstructedWithNonNilCore(r, corePtrType) {
			t.Fatalf("resource %T has nil or missing *resourcecore.Core on the constructed value", r)
		}
	}
}

func resourceConstructedWithNonNilCore(r resource.Resource, corePtrType reflect.Type) bool {
	v := reflect.ValueOf(r)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	f := v.FieldByName("Core")
	if !f.IsValid() || f.Type() != corePtrType {
		return false
	}
	return !f.IsNil()
}

func typeEmbedsCorePtr(rt reflect.Type, corePtr reflect.Type) bool {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return false
	}
	for f := range rt.Fields() {
		if f.Anonymous && f.Type == corePtr {
			return true
		}
	}
	return false
}
