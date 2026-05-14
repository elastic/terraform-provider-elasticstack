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

package kibanautil

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestConflictRetry(t *testing.T) {
	t.Run("immediate_success_returns_result_without_retry", func(t *testing.T) {
		calls := 0
		result, diags := ConflictRetry(context.Background(), ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			return "ok", http.StatusOK, nil
		})

		if result != "ok" {
			t.Fatalf("got result %q, want %q", result, "ok")
		}
		if diags.HasError() {
			t.Fatalf("unexpected error: %s", diags[0].Summary())
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("retries_on_409_then_succeeds", func(t *testing.T) {
		calls := 0
		result, diags := ConflictRetry(context.Background(), ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			if calls < 3 {
				return "", http.StatusConflict, diag.Diagnostics{diag.NewErrorDiagnostic("conflict", "write lock")}
			}
			return "ok", http.StatusOK, nil
		})

		if result != "ok" {
			t.Fatalf("got result %q, want %q", result, "ok")
		}
		if diags.HasError() {
			t.Fatalf("unexpected error: %s", diags[0].Summary())
		}
		if calls != 3 {
			t.Fatalf("fn called %d times, want 3", calls)
		}
	})

	t.Run("max_attempts_exhausted_returns_last_error", func(t *testing.T) {
		calls := 0
		_, diags := ConflictRetry(context.Background(), ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			return "", http.StatusConflict, diag.Diagnostics{diag.NewErrorDiagnostic("conflict", "write lock")}
		})

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if calls != ConflictMaxAttempts {
			t.Fatalf("fn called %d times, want %d", calls, ConflictMaxAttempts)
		}
	})

	t.Run("non_409_error_passes_through_without_retry", func(t *testing.T) {
		calls := 0
		_, diags := ConflictRetry(context.Background(), ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			return "", http.StatusInternalServerError, diag.Diagnostics{diag.NewErrorDiagnostic("server error", "500")}
		})

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("context_cancelled_during_backoff_preserves_diagnostics", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		_, diags := ConflictRetry(ctx, ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			cancel()
			return "", http.StatusConflict, diag.Diagnostics{diag.NewErrorDiagnostic("conflict", "write lock")}
		})

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if len(diags) < 2 {
			t.Fatalf("got %d diagnostics, want at least 2 (original 409 + context error)", len(diags))
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("max_attempts_1_means_no_retry", func(t *testing.T) {
		calls := 0
		_, diags := ConflictRetry(context.Background(), 1, func() (string, int, diag.Diagnostics) {
			calls++
			return "", http.StatusConflict, diag.Diagnostics{diag.NewErrorDiagnostic("conflict", "write lock")}
		})

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})

	t.Run("transport_error_status_zero_passes_through_without_retry", func(t *testing.T) {
		calls := 0
		_, diags := ConflictRetry(context.Background(), ConflictMaxAttempts, func() (string, int, diag.Diagnostics) {
			calls++
			return "", 0, diag.Diagnostics{diag.NewErrorDiagnostic("connection refused", "dial tcp: connection refused")}
		})

		if !diags.HasError() {
			t.Fatal("expected error diagnostics")
		}
		if calls != 1 {
			t.Fatalf("fn called %d times, want 1", calls)
		}
	})
}
