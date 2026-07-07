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

package sync_job_create

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/connector/syncjobget"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncjobtriggermethod"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/syncjobtype"
	"github.com/elastic/terraform-provider-elasticstack/internal/asyncutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/action"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	defaultInvokeTimeout = 30 * time.Minute
	syncJobPollInterval  = 5 * time.Second

	syncStatusCompleted = "completed"
	syncStatusError     = "error"
	syncStatusCanceled  = "canceled"
	syncStatusCancelled = "cancelled"
	syncStatusSuspended = "suspended"
)

// NewAction returns a constructor for the connector sync job create action. The
// Configure, Metadata, Schema, and Invoke prelude are owned by the
// [entitycore] action envelope; this package supplies only the schema body
// and the invoke callback.
func NewAction() action.Action {
	return entitycore.NewElasticsearchAction[Model]("connector_sync_job_create", entitycore.ElasticsearchActionOptions[Model]{
		Schema:               GetSchema,
		Invoke:               invokeCreate,
		DefaultInvokeTimeout: defaultInvokeTimeout,
	})
}

// invokeCreate is the entity-specific work for elasticstack_elasticsearch_connector_sync_job_create.
// The envelope has already decoded req.Config, resolved client, enforced version
// requirements, and applied the invoke timeout to ctx.
func invokeCreate(ctx context.Context, client *clients.ElasticsearchScopedClient, req entitycore.ActionRequest[Model]) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	model := req.Config

	params, paramsDiags := syncJobCreateParamsFromModel(model)
	diags.Append(paramsDiags...)
	if diags.HasError() {
		return diags
	}

	waitForCompletion := false
	if !model.WaitForCompletion.IsNull() && !model.WaitForCompletion.IsUnknown() {
		waitForCompletion = model.WaitForCompletion.ValueBool()
	}

	syncJobID, createDiags := esclient.CreateSyncJob(ctx, client, params.ConnectorID, params.JobType, params.TriggerMethod)
	diags.Append(createDiags...)
	if diags.HasError() {
		return diags
	}

	if !waitForCompletion {
		return diags
	}

	return waitForSyncJobCompletion(ctx, syncJobID, func(ctx context.Context, id string) (*syncjobget.Response, fwdiag.Diagnostics) {
		return esclient.GetSyncJob(ctx, client, id)
	})
}

// syncJobCreateParams holds resolved POST /_connector/_sync_job parameters.
type syncJobCreateParams struct {
	ConnectorID   string
	JobType       syncjobtype.SyncJobType
	TriggerMethod syncjobtriggermethod.SyncJobTriggerMethod
}

func syncJobCreateParamsFromModel(model Model) (syncJobCreateParams, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	connectorID := model.ConnectorID.ValueString()
	if connectorID == "" {
		diags.AddError("Invalid connector_id", "connector_id must not be empty.")
		return syncJobCreateParams{}, diags
	}

	jobTypeStr := "full"
	if !model.JobType.IsNull() && !model.JobType.IsUnknown() {
		jobTypeStr = model.JobType.ValueString()
	}

	triggerMethodStr := "on_demand"
	if !model.TriggerMethod.IsNull() && !model.TriggerMethod.IsUnknown() {
		triggerMethodStr = model.TriggerMethod.ValueString()
	}

	jobType, jobTypeDiags := parseJobType(jobTypeStr)
	diags.Append(jobTypeDiags...)
	triggerMethod, triggerDiags := parseTriggerMethod(triggerMethodStr)
	diags.Append(triggerDiags...)
	if diags.HasError() {
		return syncJobCreateParams{}, diags
	}

	return syncJobCreateParams{
		ConnectorID:   connectorID,
		JobType:       jobType,
		TriggerMethod: triggerMethod,
	}, diags
}

func parseJobType(value string) (syncjobtype.SyncJobType, fwdiag.Diagnostics) {
	switch value {
	case "full":
		return syncjobtype.Full, nil
	case "incremental":
		return syncjobtype.Incremental, nil
	case "access_control":
		return syncjobtype.Accesscontrol, nil
	default:
		var diags fwdiag.Diagnostics
		diags.AddError("Invalid job_type", fmt.Sprintf("job_type must be one of full, incremental, access_control; got %q.", value))
		return syncjobtype.SyncJobType{}, diags
	}
}

func parseTriggerMethod(value string) (syncjobtriggermethod.SyncJobTriggerMethod, fwdiag.Diagnostics) {
	switch value {
	case "on_demand":
		return syncjobtriggermethod.Ondemand, nil
	case "scheduled":
		return syncjobtriggermethod.Scheduled, nil
	default:
		var diags fwdiag.Diagnostics
		diags.AddError("Invalid trigger_method", fmt.Sprintf("trigger_method must be one of on_demand, scheduled; got %q.", value))
		return syncjobtriggermethod.SyncJobTriggerMethod{}, diags
	}
}

// syncJobGetter fetches a sync job document for polling.
type syncJobGetter func(ctx context.Context, syncJobID string) (*syncjobget.Response, fwdiag.Diagnostics)

func waitForSyncJobCompletion(ctx context.Context, syncJobID string, get syncJobGetter) fwdiag.Diagnostics {
	return waitForSyncJobCompletionWithInterval(ctx, syncJobID, get, syncJobPollInterval)
}

// waitForSyncJobCompletionWithInterval delegates the poll loop to the shared
// [asyncutils.WaitForStateTransition] helper. Sync-job terminal states carry
// rich diagnostics (cancelled, suspended, error with a server message), so the
// state checker captures those via closure variables and surfaces them after
// the wait returns. ctx-deadline errors are translated into the action's own
// timeout diagnostic which embeds the last observed status.
func waitForSyncJobCompletionWithInterval(ctx context.Context, syncJobID string, get syncJobGetter, pollInterval time.Duration) fwdiag.Diagnostics {
	var (
		lastStatus    = "unknown"
		terminalDiags fwdiag.Diagnostics
		getDiags      fwdiag.Diagnostics
	)

	stateChecker := func(ctx context.Context) (bool, error) {
		job, diags := get(ctx, syncJobID)
		if diags.HasError() {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			getDiags = diags
			return false, errSyncJobGetFailed
		}

		lastStatus = job.Status.String()
		errorField := ""
		if job.Error != nil {
			errorField = *job.Error
		}

		done, statusDiags := classifyTerminalStatus(lastStatus, errorField)
		if statusDiags.HasError() {
			terminalDiags = statusDiags
			return true, nil
		}
		return done, nil
	}

	err := asyncutils.WaitForStateTransition(ctx, "connector_sync_job", syncJobID, stateChecker, asyncutils.WithPollInterval(pollInterval))

	switch {
	case terminalDiags.HasError():
		return terminalDiags
	case errors.Is(err, errSyncJobGetFailed):
		return getDiags
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return timeoutDiagnostic(syncJobID, lastStatus)
	case err != nil:
		var diags fwdiag.Diagnostics
		diags.AddError("Sync job wait failed", err.Error())
		return diags
	default:
		return nil
	}
}

// errSyncJobGetFailed is a sentinel returned by the state checker to bail out
// of the shared poll loop while preserving the original framework diagnostics
// in a closure-captured variable.
var errSyncJobGetFailed = errors.New("sync job get failed")

func timeoutDiagnostic(syncJobID, lastStatus string) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	diags.AddError(
		"Sync job did not complete within timeout",
		fmt.Sprintf("Sync job %q last observed status: %q.", syncJobID, lastStatus),
	)
	return diags
}

// classifyTerminalStatus reports whether status is terminal and returns diagnostics
// for non-success terminal states.
func classifyTerminalStatus(status string, errorField string) (done bool, diags fwdiag.Diagnostics) {
	switch status {
	case syncStatusCompleted:
		return true, diags
	case syncStatusError:
		detail := errorField
		if detail == "" {
			detail = "no error message returned by the API"
		}
		diags.AddError("Sync job failed", fmt.Sprintf("Sync job reached status error: %s.", detail))
		return true, diags
	case syncStatusCanceled, syncStatusCancelled:
		diags.AddError("Sync job cancelled", "Sync job reached terminal status cancelled.")
		return true, diags
	case syncStatusSuspended:
		detail := "Sync job reached terminal status suspended."
		if errorField != "" {
			detail = fmt.Sprintf("Sync job reached terminal status suspended: %s.", errorField)
		}
		diags.AddError("Sync job suspended", detail)
		return true, diags
	default:
		return false, diags
	}
}
