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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// minVersionStreams reflects the Kibana version where the stream.type discriminator
// field was introduced (kibana#256682). Earlier versions of the Streams API
// (9.2.x–9.3.x) reject requests containing this field.
var minVersionStreams = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel streamModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Classic streams cannot be created via the API — they must be imported.
	if planModel.ClassicConfig != nil {
		resp.Diagnostics.AddError(
			"Classic streams cannot be created",
			"Classic streams are pre-existing Elasticsearch data streams adopted by Kibana Streams. "+
				"Use `terraform import` to manage an existing classic stream instead of creating one.\n\n"+
				fmt.Sprintf("To import: terraform import elasticstack_kibana_stream.<resource_name> '%s/%s'",
					planModel.SpaceID.ValueString(), planModel.Name.ValueString()),
		)
		return
	}

	supported, sdkDiags := r.client.EnforceMinVersion(ctx, minVersionStreams)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !supported {
		resp.Diagnostics.AddError(
			"Unsupported server version",
			fmt.Sprintf("Kibana Streams require Elastic Stack %s or later.", minVersionStreams),
		)
		return
	}

	spaceID := planModel.SpaceID.ValueString()
	name := planModel.Name.ValueString()
	compositeID := clients.CompositeID{ClusterID: spaceID, ResourceID: name}
	planModel.ID = types.StringValue(compositeID.String())

	readModel := r.upsert(ctx, planModel, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if readModel == nil {
		resp.Diagnostics.AddError("Error reading stream after creation", "The stream was created but could not be read back.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, *readModel)...)
}
