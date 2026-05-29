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
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/syncjobtriggermethod"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/syncjobtype"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSchema_attributesPresent(t *testing.T) {
	t.Parallel()

	schema := GetSchema(context.Background())
	attrs := schema.GetAttributes()

	for _, name := range []string{"connector_id", "job_type", "trigger_method", "wait_for_completion"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, ok := attrs[name]
			require.True(t, ok, "schema missing attribute %q", name)
		})
	}

	require.Contains(t, schema.MarkdownDescription, "POST /_connector/_sync_job")
}

func TestGetSchema_validatorsPresent(t *testing.T) {
	t.Parallel()

	schema := GetSchema(context.Background())
	attrs := schema.GetAttributes()

	connectorID, ok := attrs["connector_id"].(actionschema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, connectorID.Validators)
	assertStringValidatorRejects(t, connectorID.Validators, "")

	jobType, ok := attrs["job_type"].(actionschema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, jobType.Validators)
	assertStringValidatorRejects(t, jobType.Validators, "invalid")
	assertStringValidatorAccepts(t, jobType.Validators, "full")

	triggerMethod, ok := attrs["trigger_method"].(actionschema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, triggerMethod.Validators)
	assertStringValidatorRejects(t, triggerMethod.Validators, "invalid")
	assertStringValidatorAccepts(t, triggerMethod.Validators, "on_demand")
}

func TestDefaultInvokeTimeout(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 30*time.Minute, defaultInvokeTimeout)
}

func TestSyncJobCreateParamsFromModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		model           Model
		wantConnectorID string
		wantJobType     syncjobtype.SyncJobType
		wantTrigger     syncjobtriggermethod.SyncJobTriggerMethod
		wantErr         bool
	}{
		{
			name: "defaults when job_type and trigger_method null",
			model: Model{
				ConnectorID: types.StringValue("conn-1"),
			},
			wantConnectorID: "conn-1",
			wantJobType:     syncjobtype.Full,
			wantTrigger:     syncjobtriggermethod.Ondemand,
		},
		{
			name: "defaults when job_type and trigger_method unknown",
			model: Model{
				ConnectorID:   types.StringValue("conn-2"),
				JobType:       types.StringUnknown(),
				TriggerMethod: types.StringUnknown(),
			},
			wantConnectorID: "conn-2",
			wantJobType:     syncjobtype.Full,
			wantTrigger:     syncjobtriggermethod.Ondemand,
		},
		{
			name: "explicit values respected",
			model: Model{
				ConnectorID:   types.StringValue("conn-3"),
				JobType:       types.StringValue("incremental"),
				TriggerMethod: types.StringValue("scheduled"),
			},
			wantConnectorID: "conn-3",
			wantJobType:     syncjobtype.Incremental,
			wantTrigger:     syncjobtriggermethod.Scheduled,
		},
		{
			name: "access_control job type",
			model: Model{
				ConnectorID: types.StringValue("conn-4"),
				JobType:     types.StringValue("access_control"),
			},
			wantConnectorID: "conn-4",
			wantJobType:     syncjobtype.Accesscontrol,
			wantTrigger:     syncjobtriggermethod.Ondemand,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			params, diags := syncJobCreateParamsFromModel(tc.model)
			if tc.wantErr {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError(), diags)
			assert.Equal(t, tc.wantConnectorID, params.ConnectorID)
			assert.Equal(t, tc.wantJobType, params.JobType)
			assert.Equal(t, tc.wantTrigger, params.TriggerMethod)
		})
	}
}

func TestClassifyTerminalStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     string
		errorField string
		wantDone   bool
		wantErr    bool
		wantDetail string
	}{
		{
			name:     "completed",
			status:   "completed",
			wantDone: true,
		},
		{
			name:       "error with message",
			status:     "error",
			errorField: "permission denied",
			wantDone:   true,
			wantErr:    true,
			wantDetail: "permission denied",
		},
		{
			name:     "error without message",
			status:   "error",
			wantDone: true,
			wantErr:  true,
		},
		{
			name:     "canceled API spelling",
			status:   "canceled",
			wantDone: true,
			wantErr:  true,
		},
		{
			name:     "cancelled spec spelling",
			status:   "cancelled",
			wantDone: true,
			wantErr:  true,
		},
		{
			name:     "suspended",
			status:   "suspended",
			wantDone: true,
			wantErr:  true,
		},
		{
			name:     "pending",
			status:   "pending",
			wantDone: false,
		},
		{
			name:     "in_progress",
			status:   "in_progress",
			wantDone: false,
		},
		{
			name:     "canceling",
			status:   "canceling",
			wantDone: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			done, diags := classifyTerminalStatus(tc.status, tc.errorField)
			assert.Equal(t, tc.wantDone, done)
			if tc.wantErr {
				require.True(t, diags.HasError())
				if tc.wantDetail != "" {
					assert.Contains(t, diags.Errors()[0].Detail(), tc.wantDetail)
				}
				return
			}
			require.False(t, diags.HasError())
		})
	}
}

func TestModel_GetVersionRequirements(t *testing.T) {
	t.Parallel()

	reqs, diags := Model{}.GetVersionRequirements()
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	assert.Equal(t, "8.12.0", reqs[0].MinVersion.String())
	assert.Contains(t, reqs[0].ErrorMessage, "8.12.0")
}

func assertStringValidatorRejects(t *testing.T, validators []validator.String, value string) {
	t.Helper()
	var resp validator.StringResponse
	for _, v := range validators {
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(value),
		}, &resp)
	}
	require.True(t, resp.Diagnostics.HasError(), "expected validation error for value %q", value)
}

func assertStringValidatorAccepts(t *testing.T, validators []validator.String, value string) {
	t.Helper()
	var resp validator.StringResponse
	for _, v := range validators {
		v.ValidateString(context.Background(), validator.StringRequest{
			ConfigValue: types.StringValue(value),
		}, &resp)
	}
	require.False(t, resp.Diagnostics.HasError(), "expected validation success for value %q: %v", value, resp.Diagnostics)
}
