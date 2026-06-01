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

package security_entity_store

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	defaultSpaceID = "default"
	resourceID     = "entity_store"
)

var (
	_ resource.Resource                = newResource()
	_ resource.ResourceWithConfigure   = newResource()
	_ resource.ResourceWithImportState = newResource()

	MinVersion = version.Must(version.NewVersion("9.1.0"))
)

type Resource struct {
	*entitycore.KibanaResource[tfModel]
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[tfModel](
			entitycore.ComponentKibana,
			"security_entity_store",
			entitycore.KibanaResourceOptions[tfModel]{
				Schema: getSchema,
				Create: createEntityStore,
				Read:   readEntityStore,
				Update: updateEntityStore,
				Delete: deleteEntityStore,
			},
		),
	}
}

func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) == 2 && parts[1] == resourceID {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_id"), parts[0])...)
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildID(spaceID string) string {
	if spaceID == "" {
		spaceID = defaultSpaceID
	}
	return spaceID + "/" + resourceID
}

func normalizeSpaceID(v types.String) string {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return defaultSpaceID
	}
	return v.ValueString()
}
