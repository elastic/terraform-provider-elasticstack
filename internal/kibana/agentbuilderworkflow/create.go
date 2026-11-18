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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var createWorkflow = entitycore.SimpleKibanaCreate[workflowModel, kbapi.PostWorkflowsWorkflowJSONRequestBody, models.Workflow](
	func(plan workflowModel, _ context.Context) (kbapi.PostWorkflowsWorkflowJSONRequestBody, diag.Diagnostics) {
		return plan.toAPICreateModel(), nil
	},
	kibanaoapi.CreateWorkflow,
	(*workflowModel).populateWrittenCreate,
)

// populateWrittenCreate sets SpaceID explicitly so the returned model
// carries the resolved space for the envelope's read-after-write step, and
// captures workflow_id: it is Computed+Optional, and when the caller omits
// it, the API generates one and returns it on the POST response.
func (model *workflowModel) populateWrittenCreate(_ context.Context, spaceID string, created *models.Workflow) diag.Diagnostics {
	var diags diag.Diagnostics

	model.SpaceID = types.StringValue(spaceID)

	if created != nil {
		model.WorkflowID = types.StringValue(created.ID)
		if !created.Valid {
			diags.AddError("Invalid workflow", "The workflow was created but its configuration is invalid. Please check the YAML definition.")
		}
	}

	return diags
}
