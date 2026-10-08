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

package indexsettings

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = newIndexSettingsResource()
	_ resource.ResourceWithValidateConfig = newIndexSettingsResource()
	_ resource.ResourceWithImportState    = newIndexSettingsResource()
)

type indexSettingsResource struct {
	*entitycore.ElasticsearchResource[tfModel]
}

func newIndexSettingsResource() *indexSettingsResource {
	return &indexSettingsResource{
		ElasticsearchResource: entitycore.NewElasticsearchResource[tfModel]("index_settings", entitycore.ElasticsearchResourceOptions[tfModel]{
			Schema:   getSchemaFactory,
			Create:   createIndexSettings,
			Update:   updateIndexSettings,
			Read:     readIndexSettings,
			Delete:   deleteIndexSettings,
			PostRead: postReadIndexSettings,
		}),
	}
}

func NewIndexSettingsResource() resource.Resource {
	return newIndexSettingsResource()
}

// ValidateConfig runs the resource-level declared-settings validation: no
// settings_json key may overlap a configured typed dynamic attribute. An
// index-only configuration is valid; the empty settings_json object is
// rejected by the attribute validator.
func (r *indexSettingsResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config tfModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateDeclaredSettings(config)...)
}

// ImportState parses the "<cluster_uuid>/<index_name>" composite import ID,
// sets id and index, and records the import in private state so the first read
// hydrates the known dynamic typed attributes (settings_json stays unset).
// resource.ImportStatePassthroughID is insufficient here: it sets only id.
func (r *indexSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	entitycore.NewCompositeIDImporter(path.Root("id"), path.Root("index")).ImportState(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Private != nil {
		resp.Diagnostics.Append(setImportHydrationFlag(ctx, resp.Private)...)
	}
}
