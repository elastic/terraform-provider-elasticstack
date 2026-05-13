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

package anomalydetectionjob

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCalendarIDsFromTFSet_unknownOrNull(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()
		out, diags := calendarIDsFromTFSet(ctx, types.SetUnknown(types.StringType))
		if diags.HasError() {
			t.Fatal(diags)
		}
		if out != nil {
			t.Fatalf("expected nil slice, got %#v", out)
		}
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()
		out, diags := calendarIDsFromTFSet(ctx, types.SetNull(types.StringType))
		if diags.HasError() {
			t.Fatal(diags)
		}
		if out != nil {
			t.Fatalf("expected nil slice, got %#v", out)
		}
	})
}
