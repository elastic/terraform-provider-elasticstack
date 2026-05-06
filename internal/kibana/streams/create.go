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

package streams

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createStream(ctx context.Context, client *clients.KibanaScopedClient, spaceID string, plan streamModel) (streamModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Classic streams cannot be created via the API — they must be imported.
	if plan.ClassicConfig != nil {
		diags.AddError(
			"Classic streams cannot be created",
			"Classic streams are pre-existing Elasticsearch data streams adopted by Kibana Streams. "+
				"Use `terraform import` to manage an existing classic stream instead of creating one.\n\n"+
				fmt.Sprintf("To import: terraform import elasticstack_kibana_stream.<resource_name> '%s/%s'",
					spaceID, plan.GetResourceID().ValueString()),
		)
		return streamModel{}, diags
	}

	name := plan.GetResourceID().ValueString()
	compositeID := clients.CompositeID{ClusterID: spaceID, ResourceID: name}
	plan.ID = types.StringValue(compositeID.String())

	readModel, upsertDiags := upsertStream(ctx, client, plan)
	diags.Append(upsertDiags...)
	if diags.HasError() {
		return streamModel{}, diags
	}
	if readModel == nil {
		diags.AddError("Error reading stream after creation", "The stream was created but could not be read back.")
		return streamModel{}, diags
	}

	return *readModel, diags
}
