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

package templateutil

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
)

type stringValue interface {
	IsNull() bool
	IsUnknown() bool
	ValueString() string
}

// IsKnownSemanticallyEmpty reports whether a prior JSON string value is a
// known, non-null value that nevertheless decodes to a zero-length JSON object
// (for example `{}` or whitespace-padded variants). The flatten layer uses this
// signal to preserve a practitioner-authored empty-object value in state when
// the Elasticsearch GET response omits the corresponding field entirely.
func IsKnownSemanticallyEmpty(v stringValue) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}
	return typeutils.IsEmptyJSONObject(v.ValueString())
}
