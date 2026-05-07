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

package datafeed

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// deleteDatafeed stops and deletes the datafeed. It satisfies the entitycore
// elasticsearchDeleteFunc[Datafeed] signature.
func deleteDatafeed(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, _ Datafeed) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	datafeedID := resourceID
	if datafeedID == "" {
		diags.AddError("Invalid Configuration", "datafeed_id cannot be empty")
		return diags
	}

	// Before deleting, we need to stop the datafeed if it's running
	_, stopDiags := maybeStopDatafeed(ctx, client, datafeedID)
	diags.Append(stopDiags...)
	if diags.HasError() {
		return diags
	}

	// Delete the datafeed
	deleteDiags := elasticsearch.DeleteDatafeed(ctx, client, datafeedID, false)
	diags.Append(deleteDiags...)
	return diags
}

// maybeStopDatafeed stops the datafeed if it is currently running. Returns
// true if the datafeed was stopped and should be restarted after an update.
func maybeStopDatafeed(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	datafeedID string,
) (bool, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	// Check current state
	currentState, stateDiags := GetDatafeedState(ctx, client, datafeedID)
	diags.Append(stateDiags...)
	if diags.HasError() {
		return false, diags
	}

	if currentState == nil {
		return false, nil
	}

	// If the datafeed is not running, nothing to stop
	if *currentState != StateStarted && *currentState != StateStarting {
		return false, diags
	}

	// Stop the datafeed
	stopDiags := elasticsearch.StopDatafeed(ctx, client, datafeedID, false, 0)
	diags.Append(stopDiags...)
	if diags.HasError() {
		return true, diags
	}

	// Wait for the datafeed to reach stopped state
	_, waitDiags := WaitForDatafeedState(ctx, client, datafeedID, StateStopped)
	diags.Append(waitDiags...)
	if diags.HasError() {
		return true, diags
	}

	return true, diags
}
