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

package anomalydetectionjob

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// createAnomalyDetectionJob creates the ML job and sets the composite ID on the
// returned model. It satisfies the entitycore ElasticsearchCreateFunc[TFModel]
// signature. The envelope handles the read-after-write and state persistence.
func createAnomalyDetectionJob(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, plan TFModel) (TFModel, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	jobID := resourceID

	// Convert TF model to API model
	apiModel, convDiags := plan.toAPIModel(ctx)
	diags.Append(convDiags...)
	if diags.HasError() {
		return plan, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating ML anomaly detection job: %s", jobID))

	typedClient, err := client.GetESClient()
	if err != nil {
		diags.AddError("Failed to get Elasticsearch client", err.Error())
		return plan, diags
	}

	// Build typed request and call the typed API
	putReq := apiModel.toPutJobRequest()
	_, err = typedClient.Ml.PutJob(jobID).Request(&putReq).Do(ctx)
	if err != nil {
		diags.AddError("Failed to create ML anomaly detection job", fmt.Sprintf("Unable to create ML anomaly detection job: %s — %s", jobID, err.Error()))
		return plan, diags
	}

	// Set the composite ID so the envelope and readFunc can carry it through.
	compID, sdkDiags := client.ID(ctx, jobID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return plan, diags
	}

	plan.ID = types.StringValue(compID.String())

	tflog.Debug(ctx, fmt.Sprintf("Successfully created ML anomaly detection job: %s", jobID))
	return plan, diags
}
