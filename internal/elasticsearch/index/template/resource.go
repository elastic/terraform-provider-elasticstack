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

package template

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Resource implements the elasticstack_elasticsearch_index_template resource.
// It embeds *entitycore.ElasticsearchResource[Model] for schema injection, Read,
// and Delete. Create and Update remain on the concrete type because they require
// config-derived alias reconciliation, server-version gating, and the 8.x
// allow_custom_routing workaround that cannot be expressed via the callback contract.
type Resource struct {
	*entitycore.ElasticsearchResource[Model]
}

func newResource() *Resource {
	createFn, updateFn := entitycore.PlaceholderElasticsearchWriteCallbacks[Model]()
	return &Resource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[Model](
			entitycore.ComponentElasticsearch,
			"index_template",
			resourceSchema,
			readIndexTemplate,
			deleteIndexTemplate,
			createFn,
			updateFn,
		),
	}
}

func NewResource() resource.Resource {
	return newResource()
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: upgradeStateV0ToV1(),
	}
}

var (
	_ resource.Resource                   = &Resource{}
	_ resource.ResourceWithConfigure      = &Resource{}
	_ resource.ResourceWithImportState    = &Resource{}
	_ resource.ResourceWithModifyPlan     = &Resource{}
	_ resource.ResourceWithValidateConfig = &Resource{}
	_ resource.ResourceWithUpgradeState   = &Resource{}
)
