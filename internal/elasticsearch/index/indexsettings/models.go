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
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/dynamicsettings"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfModel is the Terraform model for elasticstack_elasticsearch_index_settings.
// The dynamic-setting fields come from the shared, anonymously-embedded
// dynamicsettings.Model, matching indexparent.GetDynamicSettingAttributes().
type tfModel struct {
	entitycore.ResourceTimeoutsField
	ID    types.String `tfsdk:"id"`
	Index types.String `tfsdk:"index"`
	dynamicsettings.Model
	SettingsJSON            jsontypes.Normalized `tfsdk:"settings_json"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
}

// GetID satisfies [entitycore.ElasticsearchResourceModel].
func (model tfModel) GetID() types.String { return model.ID }

// GetResourceID satisfies [entitycore.ElasticsearchResourceModel], returning the
// target concrete index name used as the write identity.
func (model tfModel) GetResourceID() types.String { return model.Index }

// GetElasticsearchConnection satisfies [entitycore.ElasticsearchResourceModel].
func (model tfModel) GetElasticsearchConnection() types.List { return model.ElasticsearchConnection }
