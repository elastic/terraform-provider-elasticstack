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

package asyncutils

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitWithGetterDiagnostics_becomesDoneAfterPolls(t *testing.T) {
	t.Parallel()

	var calls int
	get := func(_ context.Context) (int, diag.Diagnostics) {
		calls++
		return calls, nil
	}
	isDone := func(value int) (bool, diag.Diagnostics) {
		return value >= 2, nil
	}
	onWaitError := func(_ int, err error) diag.Diagnostics {
		t.Fatalf("onWaitError should not be called, got %v", err)
		return nil
	}

	last, diags := WaitWithGetterDiagnostics(context.Background(), "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(5*time.Millisecond))
	require.False(t, diags.HasError(), diags)
	assert.GreaterOrEqual(t, calls, 2)
	assert.Equal(t, calls, last)
}

func TestWaitWithGetterDiagnostics_checksImmediately(t *testing.T) {
	t.Parallel()

	var calls int
	get := func(_ context.Context) (int, diag.Diagnostics) {
		calls++
		return 42, nil
	}
	isDone := func(int) (bool, diag.Diagnostics) { return true, nil }
	onWaitError := func(_ int, err error) diag.Diagnostics {
		t.Fatalf("onWaitError should not be called, got %v", err)
		return nil
	}

	last, diags := WaitWithGetterDiagnostics(context.Background(), "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(time.Hour))
	require.False(t, diags.HasError(), diags)
	assert.Equal(t, 1, calls)
	assert.Equal(t, 42, last)
}

func TestWaitWithGetterDiagnostics_timeout(t *testing.T) {
	t.Parallel()

	get := func(_ context.Context) (string, diag.Diagnostics) {
		return "pending", nil
	}
	isDone := func(string) (bool, diag.Diagnostics) { return false, nil }
	onWaitError := func(last string, err error) diag.Diagnostics {
		var diags diag.Diagnostics
		diags.AddError("timed out", last+": "+err.Error())
		return diags
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	last, diags := WaitWithGetterDiagnostics(ctx, "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(5*time.Millisecond))
	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "timed out", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "pending")
	assert.Equal(t, "pending", last)
}

func TestWaitWithGetterDiagnostics_propagatesGetDiagnostics(t *testing.T) {
	t.Parallel()

	wantSummary := "getter exploded"
	wantDetail := "permission denied"
	get := func(_ context.Context) (string, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError(wantSummary, wantDetail)
		return "", diags
	}
	isDone := func(string) (bool, diag.Diagnostics) { return false, nil }
	onWaitError := func(_ string, err error) diag.Diagnostics {
		t.Fatalf("onWaitError should not be called, got %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, diags := WaitWithGetterDiagnostics(ctx, "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(5*time.Millisecond))
	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, wantSummary, diags.Errors()[0].Summary())
	assert.Equal(t, wantDetail, diags.Errors()[0].Detail())
}

func TestWaitWithGetterDiagnostics_terminalErrorState(t *testing.T) {
	t.Parallel()

	get := func(_ context.Context) (string, diag.Diagnostics) {
		return "failed", nil
	}
	isDone := func(value string) (bool, diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("job failed", "reached terminal failure status")
		return true, diags
	}
	onWaitError := func(_ string, err error) diag.Diagnostics {
		t.Fatalf("onWaitError should not be called, got %v", err)
		return nil
	}

	last, diags := WaitWithGetterDiagnostics(context.Background(), "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(5*time.Millisecond))
	require.True(t, diags.HasError())
	require.Len(t, diags.Errors(), 1)
	assert.Equal(t, "job failed", diags.Errors()[0].Summary())
	assert.Equal(t, "failed", last)
}

func TestWaitWithGetterDiagnostics_contextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	get := func(_ context.Context) (string, diag.Diagnostics) {
		cancel()
		return "pending", nil
	}
	isDone := func(string) (bool, diag.Diagnostics) { return false, nil }

	var gotErr error
	onWaitError := func(_ string, err error) diag.Diagnostics {
		gotErr = err
		var diags diag.Diagnostics
		diags.AddError("canceled", err.Error())
		return diags
	}

	_, diags := WaitWithGetterDiagnostics(ctx, "widget", "widget-1", get, isDone, onWaitError, WithPollInterval(5*time.Millisecond))
	require.True(t, diags.HasError())
	require.ErrorIs(t, gotErr, context.Canceled)
}
