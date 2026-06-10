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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ImportStateWithNameAttribute handles ImportState for resources that use a
// composite ID of the form "<cluster_uuid>/<name>" and expose a "name"
// attribute in their schema.
//
// When req.ID is a composite ID, the full composite value is passed through as
// the state "id" and the bare resource name is written to the "name" attribute.
// When req.ID is a plain name (no composite prefix), it is written directly to
// "name" with no "id" passthrough.
func ImportStateWithNameAttribute(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if compID, diags := clients.CompositeIDFromStr(req.ID); !diags.HasError() {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), compID.ResourceID)...)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
