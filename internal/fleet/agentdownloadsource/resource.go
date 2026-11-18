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

package agentdownloadsource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()
	_ resource.ResourceWithModifyPlan  = newResource()
)

// Resource implements the Fleet Agent Download Source resource.
type Resource struct {
	*entitycore.KibanaResource[model]
	*entitycore.SpaceImporter
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[model](
			entitycore.ComponentFleet,
			"agent_download_source",
			entitycore.KibanaResourceOptions[model]{
				Schema: getSchema,
				Read:   readAgentDownloadSource,
				Delete: deleteAgentDownloadSource,
				Create: writeAgentDownloadSource,
				Update: writeAgentDownloadSource,
			},
		),
		SpaceImporter: entitycore.NewSpaceImporter(path.Root("source_id"), path.Root("id")),
	}
}

// Delete removes the download source, then clears the stored write-only secret
// hashes so a re-created resource starts without drift state.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.KibanaResource.Delete(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	clearAllSecretHashes(ctx, resp.Private, &resp.Diagnostics)
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}
