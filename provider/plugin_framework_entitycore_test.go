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

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPluginFrameworkResourcesEmbedEntityCoreResourceBase(t *testing.T) {
	t.Parallel()

	// Match the real registration path in (*Provider).Resources, including
	// experimental resources (same branch as acceptance tests via AccTestVersion).
	p := &Provider{version: AccTestVersion}
	ctx := context.Background()

	resourceBasePtrType := reflect.TypeFor[*entitycore.ResourceBase]()

	for _, newRes := range p.Resources(ctx) {
		r := newRes()
		rt := reflect.TypeOf(r)
		if !typeEmbedsResourceBasePtr(rt, resourceBasePtrType) {
			t.Fatalf("resource %T does not embed *entitycore.ResourceBase", r)
		}
		if !resourceConstructedWithNonNilResourceBase(r, resourceBasePtrType) {
			t.Fatalf("resource %T has nil or missing *entitycore.ResourceBase on the constructed value", r)
		}
	}
}

func resourceConstructedWithNonNilResourceBase(r resource.Resource, resourceBasePtrType reflect.Type) bool {
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
	f := v.FieldByName("ResourceBase")
	if !f.IsValid() || f.Type() != resourceBasePtrType {
		return false
	}
	return !f.IsNil()
}

func typeEmbedsResourceBasePtr(rt reflect.Type, resourceBasePtr reflect.Type) bool {
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return false
	}
	for f := range rt.Fields() {
		if f.Anonymous && f.Type == resourceBasePtr {
			return true
		}
	}
	return false
}
