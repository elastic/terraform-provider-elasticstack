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
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newDatafeedResource()
	_ resource.ResourceWithConfigure   = newDatafeedResource()
	_ resource.ResourceWithImportState = newDatafeedResource()
)

type datafeedResource struct {
	*entitycore.ElasticsearchResource[Datafeed]
}

func newDatafeedResource() *datafeedResource {
	return &datafeedResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource(
			entitycore.ComponentElasticsearch,
			"ml_datafeed",
			getSchema,
			readDatafeed,
			deleteDatafeed,
			createDatafeed,
			updateDatafeed,
		),
	}
}

func NewDatafeedResource() resource.Resource {
	return newDatafeedResource()
}

func (r *datafeedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	compID, sdkDiags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	datafeedID := compID.ResourceID

	// Set the datafeed_id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("datafeed_id"), datafeedID)...)
}
