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

package agentbuilderworkflow

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createWorkflow(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[workflowModel]) (entitycore.KibanaWriteResult[workflowModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body := plan.toAPICreateModel()

	oapiClient := client.GetKibanaOapiClient()

	created, d := kibanaoapi.CreateWorkflow(ctx, oapiClient, req.SpaceID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[workflowModel]{}, diags
	}

	plan.SpaceID = types.StringValue(req.SpaceID)
	// workflow_id is Computed+Optional: when the caller omits it, the API
	// generates one and returns it on the POST response. Capture it on the
	// plan so the envelope's read-after-write step can resolve the identity.
	if created != nil {
		plan.WorkflowID = types.StringValue(created.ID)
		if !created.Valid {
			diags.AddError("Invalid workflow", "The workflow was created but its configuration is invalid. Please check the YAML definition.")
		}
	}

	return entitycore.KibanaWriteResult[workflowModel]{Model: plan}, diags
}

func readWorkflow(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, prior workflowModel) (workflowModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Restore space_id so populateFromAPI can build the composite ID correctly.
	prior.SpaceID = types.StringValue(spaceID)

	oapiClient := client.GetKibanaOapiClient()

	workflow, d := kibanaoapi.GetWorkflow(ctx, oapiClient, spaceID, resourceID)
	diags.Append(d...)
	if diags.HasError() {
		return prior, false, diags
	}

	if workflow == nil {
		return prior, false, diags
	}

	prior.populateFromAPI(workflow)
	return prior, true, diags
}

func updateWorkflow(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[workflowModel]) (entitycore.KibanaWriteResult[workflowModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	body := plan.toAPIUpdateModel()

	oapiClient := client.GetKibanaOapiClient()

	updated, d := kibanaoapi.UpdateWorkflow(ctx, oapiClient, req.SpaceID, req.WriteID, body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[workflowModel]{}, diags
	}

	// SpaceID is set explicitly so the returned model carries the resolved
	// space for the envelope's read-after-write step.
	plan.SpaceID = types.StringValue(req.SpaceID)

	if updated != nil && !updated.Valid {
		diags.AddError("Invalid workflow", "The workflow was updated but its configuration is invalid. Please check the YAML definition.")
	}

	return entitycore.KibanaWriteResult[workflowModel]{Model: plan}, diags
}

func deleteWorkflow(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, _ workflowModel) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()
	return kibanaoapi.DeleteWorkflow(ctx, oapiClient, spaceID, resourceID)
}
