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

package datafeed

import (
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Test_durationPointerToString(t *testing.T) {
	t.Run("nil pointer returns StringNull", func(t *testing.T) {
		result, err := durationPointerToString(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsNull() {
			t.Errorf("expected null, got %v", result)
		}
	})

	t.Run("string Duration returns StringValue", func(t *testing.T) {
		d := estypes.Duration("10m")
		result, err := durationPointerToString(&d)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsNull() || result.IsUnknown() {
			t.Fatalf("expected non-null value, got null/unknown")
		}
		if result.ValueString() != "10m" {
			t.Errorf("expected '10m', got %q", result.ValueString())
		}
	})

	t.Run("marshallable value returns StringValue", func(t *testing.T) {
		// A plain string marshals to a JSON-quoted string; Unmarshal handles the unquoting.
		result, err := durationPointerToString("30s")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ValueString() != "30s" {
			t.Errorf("expected '30s', got %q", result.ValueString())
		}
	})

	t.Run("non-string json value uses fallback raw bytes", func(t *testing.T) {
		// An integer marshals as a JSON number; Unmarshal into string fails,
		// so the fallback path (string(b)) is used.
		result, err := durationPointerToString(42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ValueString() != "42" {
			t.Errorf("expected '42', got %q", result.ValueString())
		}
	})

	t.Run("empty string returns empty StringValue", func(t *testing.T) {
		result, err := durationPointerToString("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := types.StringValue("")
		if result != expected {
			t.Errorf("expected empty StringValue, got %v", result)
		}
	})
}
