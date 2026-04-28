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

package ilm

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = newResource()
	_ resource.ResourceWithConfigure      = newResource()
	_ resource.ResourceWithImportState    = newResource()
	_ resource.ResourceWithValidateConfig = newResource()
	_ resource.ResourceWithUpgradeState   = newResource()
)

type Resource struct {
	*entitycore.ResourceBase
}

func newResource() *Resource {
	return &Resource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentElasticsearch, "index_lifecycle"),
	}
}

func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: migrateILMStateV0ToV1,
		},
	}
}
