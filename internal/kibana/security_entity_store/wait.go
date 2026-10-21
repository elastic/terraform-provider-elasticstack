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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type entityStoreStatusValue string

const (
	entityStoreStatusNotInstalled entityStoreStatusValue = "not_installed"
	entityStoreStatusInstalling   entityStoreStatusValue = "installing"
)

type entityStoreEngineComponentStatus struct {
	ID        string  `json:"id"`
	Installed bool    `json:"installed"`
	Resource  string  `json:"resource"`
	Health    *string `json:"health,omitempty"`
}

type entityStoreEngineStatus string

const entityStoreEngineStatusStarted entityStoreEngineStatus = "started"

type entityStoreEngineType string

// entityStoreStatus mirrors the JSON shape returned by GET /api/security/entity_store/status.
type entityStoreStatus struct {
	Status  entityStoreStatusValue `json:"status"`
	Engines []entityStoreEngine    `json:"engines"`
}

type entityStoreEngine struct {
	Components         *[]entityStoreEngineComponentStatus `json:"components,omitempty"`
	Delay              *string                             `json:"delay,omitempty"`
	DocsPerSecond      *int                                `json:"docsPerSecond,omitempty"`
	Error              *entityStoreEngineError             `json:"error,omitempty"`
	FieldHistoryLength int                                 `json:"fieldHistoryLength"`
	Filter             *string                             `json:"filter,omitempty"`
	Frequency          *string                             `json:"frequency,omitempty"`
	IndexPattern       string                              `json:"indexPattern"`
	LookbackPeriod     *string                             `json:"lookbackPeriod,omitempty"`
	Status             entityStoreEngineStatus             `json:"status"`
	Timeout            *string                             `json:"timeout,omitempty"`
	TimestampField     *string                             `json:"timestampField,omitempty"`
	Type               entityStoreEngineType               `json:"type"`
}

type entityStoreEngineError struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

func getEntityStoreStatus(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, includeComponents bool) (*entityStoreStatus, []byte, diag.Diagnostics) {
	resp, diags := kibanaoapi.GetSecurityEntityStoreStatus(ctx, client.GetKibanaOapiClient(), spaceID, includeComponents)
	if diags.HasError() {
		return nil, nil, diags
	}

	var status entityStoreStatus
	if err := json.Unmarshal(resp.Body, &status); err != nil {
		return nil, nil, diagutil.FrameworkDiagFromError(err)
	}
	return &status, resp.Body, nil
}

// waitForUninstall polls until the entity store reports it is no longer
// installed. It is bounded by the Delete ctx deadline (carried in from the
// resource timeouts block); no separate timeout is imposed.
func waitForUninstall(ctx context.Context, client *clients.KibanaScopedClient, spaceID string) diag.Diagnostics {
	getStatus := func(ctx context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
		return getEntityStoreStatus(ctx, client, spaceID, false)
	}
	checker := makeUninstallStateChecker(getStatus)

	// Immediate first check so an already not_installed store returns without
	// waiting a poll interval. Skipped when ctx is already done so the wait
	// loop reports the deadline.
	if ctx.Err() == nil {
		if done, _ := checker(ctx); done {
			return nil
		}
	}

	err := asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker, asyncutils.WithPollInterval(5*time.Second))
	return uninstallWaitDiagsFromError(err)
}

// makeUninstallStateChecker builds a StateChecker that returns true once the
// entity store status reaches "not_installed". Errors reading the status are
// treated as transient and retried.
type entityStoreStatusFunc func(ctx context.Context) (*entityStoreStatus, []byte, diag.Diagnostics)

func makeUninstallStateChecker(getStatus entityStoreStatusFunc) asyncutils.StateChecker {
	return func(ctx context.Context) (bool, error) {
		status, _, diags := getStatus(ctx)
		if diags.HasError() {
			tflog.Debug(ctx, fmt.Sprintf("transient error reading entity store status during uninstall wait: %v", diags.Errors()))
			return false, nil
		}
		return status.Status == entityStoreStatusNotInstalled, nil
	}
}

// uninstallWaitDiagsFromError converts a WaitForStateTransition error into a
// clear error diagnostic. A nil error yields nil diagnostics.
func uninstallWaitDiagsFromError(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Security Entity Store uninstall did not complete within the Delete timeout",
			err.Error(),
		),
	}
}

// waitForStarted performs an initial synchronous status read and, if the store
// is still installing, polls until it leaves the installing state (running,
// stopped, error, or not_installed). It is bounded by the Read ctx deadline
// (carried in from the resource timeouts block); no separate timeout is
// imposed. If the deadline is exceeded while still installing, it returns the
// last-observed status together with a warning diagnostic so that Read can
// proceed with a degraded read rather than failing.
func waitForStarted(ctx context.Context, client *clients.KibanaScopedClient, spaceID string) (*entityStoreStatus, []byte, diag.Diagnostics) {
	return waitForStartedFromStatusFunc(ctx, func(ctx context.Context) (*entityStoreStatus, []byte, diag.Diagnostics) {
		return getEntityStoreStatus(ctx, client, spaceID, false)
	}, spaceID, asyncutils.WithPollInterval(3*time.Second))
}

// waitForStartedFromStatusFunc is the internal implementation of waitForStarted,
// exposed with a swappable status reader so unit tests can exercise it without a
// real Kibana client.
func waitForStartedFromStatusFunc(ctx context.Context, getStatus entityStoreStatusFunc, spaceID string, opts ...asyncutils.Option) (*entityStoreStatus, []byte, diag.Diagnostics) {
	status, rawBody, diags := getStatus(ctx)
	if diags.HasError() {
		return nil, nil, diags
	}
	if status.Status != entityStoreStatusInstalling {
		return status, rawBody, nil
	}

	checker := makeStartedStateChecker(getStatus, func(s *entityStoreStatus, b []byte) {
		status, rawBody = s, b
	})

	err := asyncutils.WaitForStateTransition(ctx, "security entity store", spaceID, checker, opts...)
	return status, rawBody, startedWaitDiagsFromError(err)
}

// makeStartedStateChecker builds a StateChecker that returns true once the entity
// store status is no longer "installing" (i.e. running, stopped, error, or
// not_installed). Errors reading the status are treated as transient and retried.
// The capture callback, if non-nil, receives the latest successful status read so
// the caller can return the last-observed data on timeout.
func makeStartedStateChecker(getStatus entityStoreStatusFunc, capture func(*entityStoreStatus, []byte)) asyncutils.StateChecker {
	return func(ctx context.Context) (bool, error) {
		status, rawBody, diags := getStatus(ctx)
		if diags.HasError() {
			tflog.Debug(ctx, fmt.Sprintf("transient error reading entity store status during started wait: %v", diags.Errors()))
			return false, nil
		}
		if capture != nil {
			capture(status, rawBody)
		}
		return status.Status != entityStoreStatusInstalling, nil
	}
}

// startedWaitDiagsFromError converts a WaitForStateTransition error into diagnostics.
// A nil error yields nil diagnostics. Only a Read timeout (context deadline
// exceeded) is downgraded to a warning so Read can degrade to a partial read;
// any other error (including context cancellation or an unexpected wait failure)
// is returned as an error diagnostic so the read fails fast rather than
// proceeding after cancellation.
func startedWaitDiagsFromError(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return diag.Diagnostics{
			diag.NewWarningDiagnostic(
				"Security Entity Store is still installing; returning partial read data",
				err.Error(),
			),
		}
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Failed while waiting for Security Entity Store to finish installing",
			err.Error(),
		),
	}
}
