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

package entitycore

import (
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
)

const (
	DefaultResourceCreateTimeout = 20 * time.Minute
	DefaultResourceReadTimeout   = 5 * time.Minute
	DefaultResourceUpdateTimeout = 20 * time.Minute
	DefaultResourceDeleteTimeout = 20 * time.Minute
)

// ResourceTimeoutsField is an embeddable struct that provides the resource
// `timeouts` attribute for models used with [NewElasticsearchResource] or
// [NewKibanaResource]. Embedding it satisfies [WithResourceTimeouts] without
// requiring the concrete model to redeclare the framework type.
type ResourceTimeoutsField struct {
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

// GetTimeouts returns the timeouts attribute value.
func (f ResourceTimeoutsField) GetTimeouts() timeouts.Value {
	return f.Timeouts
}

// WithResourceTimeouts is the timeouts portion of the resource model contract.
// Concrete resource models satisfy it by embedding [ResourceTimeoutsField] (or
// by declaring an equivalent field plus method).
type WithResourceTimeouts interface {
	GetTimeouts() timeouts.Value
}

// ResourceTimeouts holds per-operation default durations passed via
// [ElasticsearchResourceOptions] or [KibanaResourceOptions]. Each field that is
// zero falls back to the matching package constant at envelope call sites:
// [DefaultResourceCreateTimeout], [DefaultResourceReadTimeout],
// [DefaultResourceUpdateTimeout], or [DefaultResourceDeleteTimeout].
type ResourceTimeouts struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}

// preserveModelTimeouts copies the envelope-owned timeouts value onto a callback-
// returned model before State.Set so conversion succeeds when callbacks reconstruct
// the struct without ResourceTimeoutsField (zero timeouts.Value{} / Object[]).
func preserveModelTimeouts(model any, value timeouts.Value) {
	if model == nil {
		return
	}
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	}
	el := v.Elem()
	if f := el.FieldByName("ResourceTimeoutsField"); f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(ResourceTimeoutsField{Timeouts: value}))
		return
	}
	if f := el.FieldByName("Timeouts"); f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(value))
	}
}
