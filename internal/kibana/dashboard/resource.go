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

package dashboard

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.Resource                 = newResource()
	_ resource.ResourceWithConfigure    = newResource()
	_ resource.ResourceWithImportState  = newResource()
	_ resource.ResourceWithUpgradeState = newResource()
)

type Resource struct {
	*entitycore.KibanaResource[models.DashboardModel]
}

func schemaFactory(_ context.Context) rschema.Schema {
	return getSchema()
}

func newResource() *Resource {
	return &Resource{
		KibanaResource: entitycore.NewKibanaResource[models.DashboardModel](
			entitycore.ComponentKibana,
			"dashboard",
			entitycore.KibanaResourceOptions[models.DashboardModel]{
				Schema:   schemaFactory,
				Read:     readDashboard,
				Delete:   deleteDashboard,
				Create:   createDashboard,
				Update:   updateDashboard,
				PostRead: postReadDashboard,
			},
		),
	}
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
