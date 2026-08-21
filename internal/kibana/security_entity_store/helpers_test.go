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

package security_entity_store

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUninstallWaitDiagsFromError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		giveErr     error
		wantError   bool
		wantSummary string
	}{
		{
			name:    "nil error returns no diagnostics",
			giveErr: nil,
		},
		{
			name:        "context deadline exceeded maps to error diagnostic",
			giveErr:     context.DeadlineExceeded,
			wantError:   true,
			wantSummary: "Security Entity Store uninstall did not complete within the Delete timeout",
		},
		{
			name:        "arbitrary error maps to error diagnostic",
			giveErr:     errors.New("something failed"),
			wantError:   true,
			wantSummary: "Security Entity Store uninstall did not complete within the Delete timeout",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			diags := uninstallWaitDiagsFromError(tc.giveErr)
			if !tc.wantError {
				assert.False(t, diags.HasError())
				return
			}
			require.True(t, diags.HasError())
			assert.Equal(t, tc.wantSummary, diags.Errors()[0].Summary())
			assert.Equal(t, tc.giveErr.Error(), diags.Errors()[0].Detail())
		})
	}
}

func TestMakeUninstallFetch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		statusFunc entityStoreStatusFunc
		wantStatus entityStoreStatusValue
		wantNil    bool
	}{
		{
			name: "status read error is treated as transient retry",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("transient", "boom")
				return nil, nil, diags
			},
			wantNil: true,
		},
		{
			name: "not_installed is returned as-is",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: entityStoreStatusNotInstalled}, nil, nil
			},
			wantStatus: entityStoreStatusNotInstalled,
		},
		{
			name: "installing is returned as-is",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: entityStoreStatusInstalling}, nil, nil
			},
			wantStatus: entityStoreStatusInstalling,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fetch := makeUninstallFetch(tc.statusFunc)
			status, err := fetch(context.Background())
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, status)
				return
			}
			require.NotNil(t, status)
			assert.Equal(t, tc.wantStatus, status.Status)
		})
	}
}

func TestUninstallIsDesired(t *testing.T) {
	t.Parallel()

	assert.False(t, uninstallIsDesired(nil))
	assert.False(t, uninstallIsDesired(&entityStoreStatus{Status: entityStoreStatusInstalling}))
	assert.False(t, uninstallIsDesired(&entityStoreStatus{Status: entityStoreStatusValue("running")}))
	assert.True(t, uninstallIsDesired(&entityStoreStatus{Status: entityStoreStatusNotInstalled}))
}

func TestWaitForUninstall_DeadlineExpired(t *testing.T) {
	t.Parallel()

	// Use an already-expired context so WaitForStateTransition returns
	// immediately without calling the status checker (which lets this unit test
	// avoid a real Kibana client).
	ctx, cancel := context.WithTimeout(context.Background(), -time.Millisecond)
	defer cancel()

	diags := waitForUninstall(ctx, nil, "default")
	require.True(t, diags.HasError(), "expected an error diagnostic for an expired context")
	assert.Equal(t, "Security Entity Store uninstall did not complete within the Delete timeout", diags.Errors()[0].Summary())
	assert.Contains(t, diags.Errors()[0].Detail(), "context deadline exceeded")
}

func TestStartedWaitDiagsFromError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		giveErr     error
		wantNone    bool
		wantWarning bool
		wantError   bool
		wantSummary string
	}{
		{
			name:     "nil error returns no diagnostics",
			giveErr:  nil,
			wantNone: true,
		},
		{
			name:        "context deadline exceeded maps to warning diagnostic",
			giveErr:     context.DeadlineExceeded,
			wantWarning: true,
			wantSummary: "Security Entity Store is still installing; returning partial read data",
		},
		{
			name:        "wrapped deadline exceeded maps to warning diagnostic",
			giveErr:     fmt.Errorf("wait failed: %w", context.DeadlineExceeded),
			wantWarning: true,
			wantSummary: "Security Entity Store is still installing; returning partial read data",
		},
		{
			name:        "context canceled maps to error diagnostic",
			giveErr:     context.Canceled,
			wantError:   true,
			wantSummary: "Failed while waiting for Security Entity Store to finish installing",
		},
		{
			name:        "arbitrary error maps to error diagnostic",
			giveErr:     errors.New("something failed"),
			wantError:   true,
			wantSummary: "Failed while waiting for Security Entity Store to finish installing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			diags := startedWaitDiagsFromError(tc.giveErr)
			if tc.wantNone {
				assert.False(t, diags.HasError())
				assert.Empty(t, diags.Warnings())
				return
			}
			require.Len(t, diags, 1)
			if tc.wantWarning {
				assert.False(t, diags.HasError())
				require.Len(t, diags.Warnings(), 1)
			} else {
				require.True(t, tc.wantError)
				assert.True(t, diags.HasError())
			}
			assert.Equal(t, tc.wantSummary, diags[0].Summary())
			assert.Equal(t, tc.giveErr.Error(), diags[0].Detail())
		})
	}
}

func TestMakeStartedFetch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		statusFunc entityStoreStatusFunc
		wantNil    bool
		wantStatus entityStoreStatusValue
		wantBody   string
	}{
		{
			name: "status read error is treated as transient retry",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("transient", "boom")
				return nil, nil, diags
			},
			wantNil: true,
		},
		{
			name: "installing is returned as-is with its raw body",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: entityStoreStatusInstalling}, []byte(`{"status":"installing"}`), nil
			},
			wantStatus: entityStoreStatusInstalling,
			wantBody:   `{"status":"installing"}`,
		},
		{
			name: "running is returned as-is with its raw body",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: entityStoreStatusValue("running")}, []byte(`{"status":"running"}`), nil
			},
			wantStatus: entityStoreStatusValue("running"),
			wantBody:   `{"status":"running"}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fetch := makeStartedFetch(tc.statusFunc)
			snap, err := fetch(context.Background())
			require.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, snap)
				return
			}
			require.NotNil(t, snap)
			assert.Equal(t, tc.wantStatus, snap.status.Status)
			assert.JSONEq(t, tc.wantBody, string(snap.rawBody))
		})
	}
}

func TestStartedIsDesired(t *testing.T) {
	t.Parallel()

	assert.False(t, startedIsDesired(nil))
	assert.False(t, startedIsDesired(&entityStoreSnapshot{status: &entityStoreStatus{Status: entityStoreStatusInstalling}}))
	assert.True(t, startedIsDesired(&entityStoreSnapshot{status: &entityStoreStatus{Status: entityStoreStatusValue("running")}}))
	assert.True(t, startedIsDesired(&entityStoreSnapshot{status: &entityStoreStatus{Status: entityStoreStatusValue("stopped")}}))
	assert.True(t, startedIsDesired(&entityStoreSnapshot{status: &entityStoreStatus{Status: entityStoreStatusValue("error")}}))
	assert.True(t, startedIsDesired(&entityStoreSnapshot{status: &entityStoreStatus{Status: entityStoreStatusNotInstalled}}))
}

func TestWaitForStarted_NotInstalledEarlyExit(t *testing.T) {
	t.Parallel()

	getStatus := func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
		return &entityStoreStatus{Status: entityStoreStatusNotInstalled}, []byte(`{"status":"not_installed"}`), nil
	}

	status, rawBody, diags := waitForStartedFromStatusFunc(context.Background(), getStatus, "default")
	require.False(t, diags.HasError())
	assert.Empty(t, diags.Warnings())
	assert.Equal(t, entityStoreStatusNotInstalled, status.Status)
	assert.JSONEq(t, `{"status":"not_installed"}`, string(rawBody))
}

func TestWaitForStarted_InstallingToRunning(t *testing.T) {
	t.Parallel()

	callCount := 0
	getStatus := func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
		callCount++
		if callCount == 1 {
			return &entityStoreStatus{Status: entityStoreStatusInstalling}, []byte(`{"status":"installing"}`), nil
		}
		return &entityStoreStatus{Status: entityStoreStatusValue("running")}, []byte(`{"status":"running"}`), nil
	}

	status, rawBody, diags := waitForStartedFromStatusFunc(context.Background(), getStatus, "default", asyncutils.WithPollInterval(time.Millisecond))
	require.False(t, diags.HasError())
	assert.Empty(t, diags.Warnings())
	assert.Equal(t, entityStoreStatusValue("running"), status.Status)
	assert.JSONEq(t, `{"status":"running"}`, string(rawBody))
}

func TestWaitForStarted_DeadlineExpired(t *testing.T) {
	t.Parallel()

	getStatus := func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
		return &entityStoreStatus{Status: entityStoreStatusInstalling}, []byte(`{"status":"installing"}`), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), -time.Millisecond)
	defer cancel()

	status, rawBody, diags := waitForStartedFromStatusFunc(ctx, getStatus, "default")
	require.False(t, diags.HasError(), "expected a warning, not an error")
	require.Len(t, diags, 1)
	assert.Equal(t, diag.SeverityWarning, diags[0].Severity())
	assert.Equal(t, "Security Entity Store is still installing; returning partial read data", diags[0].Summary())
	assert.Contains(t, diags[0].Detail(), "context deadline exceeded")
	assert.Equal(t, entityStoreStatusInstalling, status.Status)
	assert.JSONEq(t, `{"status":"installing"}`, string(rawBody))
}
