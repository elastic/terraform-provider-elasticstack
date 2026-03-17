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

package exportagentbuilder

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// addInternalOriginHeader adds the required x-elastic-internal-origin header for workflow APIs
func addInternalOriginHeader(_ context.Context, req *http.Request) error {
	req.Header.Set("x-elastic-internal-origin", "Kibana")
	return nil
}

// WorkflowModel maps workflow data
type WorkflowModel struct {
	ID   types.String `tfsdk:"id"`
	Yaml types.String `tfsdk:"yaml"`
}

// FetchWorkflow fetches and parses a workflow by ID
func FetchWorkflow(ctx context.Context, client *kbapi.ClientWithResponses, workflowID string, diagnostics *diag.Diagnostics) *WorkflowModel {
	workflowResp, err := client.GetWorkflowsIdWithResponse(ctx, workflowID, addInternalOriginHeader)
	if err != nil {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: %v", workflowID, err))
		return nil
	}

	if workflowResp.StatusCode() != http.StatusOK {
		diagnostics.AddWarning("Workflow fetch failed", fmt.Sprintf("Unable to get workflow %s: HTTP %d", workflowID, workflowResp.StatusCode()))
		return nil
	}

	if workflowResp.JSON200 == nil {
		diagnostics.AddWarning("Workflow parse failed", fmt.Sprintf("Workflow %s returned nil data", workflowID))
		return nil
	}

	return &WorkflowModel{
		ID:   types.StringValue(workflowResp.JSON200.Id),
		Yaml: types.StringValue(workflowResp.JSON200.Yaml),
	}
}
