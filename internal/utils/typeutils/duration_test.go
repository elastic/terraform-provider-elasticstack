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

package typeutils

import (
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestElasticsearchDurationToString(t *testing.T) {
	t.Run("nil returns StringNull", func(t *testing.T) {
		result := ElasticsearchDurationToString(nil)
		if !result.IsNull() {
			t.Errorf("expected null, got %v", result)
		}
	})

	t.Run("string duration returns StringValue", func(t *testing.T) {
		result := ElasticsearchDurationToString(estypes.Duration("10m"))
		if result.IsNull() || result.IsUnknown() {
			t.Fatalf("expected non-null value")
		}
		if result.ValueString() != "10m" {
			t.Errorf("expected '10m', got %q", result.ValueString())
		}
	})

	t.Run("empty string returns empty StringValue", func(t *testing.T) {
		result := ElasticsearchDurationToString(estypes.Duration(""))
		expected := types.StringValue("")
		if result != expected {
			t.Errorf("expected empty StringValue, got %v", result)
		}
	})
}
