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
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
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

func TestMakeUninstallStateChecker(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		statusFunc  entityStoreStatusFunc
		wantDone    bool
		wantCheckerErr bool
	}{
		{
			name: "status read error is treated as transient retry",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("transient", "boom")
				return nil, nil, diags
			},
			wantDone:    false,
			wantCheckerErr: false,
		},
		{
			name: "not_installed reaches desired state",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: kbapi.SecurityEntityAnalyticsAPIStoreStatusNotInstalled}, nil, nil
			},
			wantDone: true,
		},
		{
			name: "installing continues polling",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: kbapi.SecurityEntityAnalyticsAPIStoreStatusInstalling}, nil, nil
			},
			wantDone: false,
		},
		{
			name: "running continues polling",
			statusFunc: func(_ context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
				return &entityStoreStatus{Status: kbapi.SecurityEntityAnalyticsAPIStoreStatusRunning}, nil, nil
			},
			wantDone: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checker := makeUninstallStateChecker(tc.statusFunc)
			done, err := checker(context.Background())
			if tc.wantCheckerErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantDone, done)
		})
	}
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
