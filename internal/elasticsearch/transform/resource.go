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

package transform

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                 = newTransformResource()
	_ resource.ResourceWithConfigure    = newTransformResource()
	_ resource.ResourceWithImportState  = newTransformResource()
	_ resource.ResourceWithUpgradeState = newTransformResource()
)

// transformResource wraps the entitycore envelope. Create and Update are
// implemented via the shared writeTransform WriteFunc (no method receiver
// overrides); UpgradeState and ImportState live on this concrete type.
type transformResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newTransformResource() *transformResource {
	return &transformResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel]("transform", entitycore.ElasticsearchResourceOptions[tfModel]{
			Schema: getSchema,
			Read:   readTransform,
			Delete: deleteTransform,
			Create: createTransform,
			Update: writeTransform,
		}),
	}
}

// NewTransformResource returns the PF resource factory for registration.
func NewTransformResource() resource.Resource {
	return newTransformResource()
}

// UpgradeState provides state upgraders for prior schema versions. The v0→v1
// upgrade unwraps singleton-list nested blocks (source, destination,
// retention_policy, sync, and their inner time blocks) into single objects
// after the schema migration from ListNestedBlock to SingleNestedBlock.
func (r *transformResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: migrateStateV0ToV1},
	}
}

// ImportState implements passthrough import on the composite id attribute.
// It also extracts the transform name from the composite ID so Read can use it.
func (r *transformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	compID, sdkDiags := clients.CompositeIDFromStr(req.ID)
	resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), compID.ResourceID)...)
}
