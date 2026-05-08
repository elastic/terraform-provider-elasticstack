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

package settings

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// resourceID is the constant write identity used for the singleton
// cluster-settings resource. It is also the ResourceID portion of the
// composite ID returned by client.ID.
const resourceID = "cluster-settings"

// tfModel is the top-level Terraform model for elasticstack_elasticsearch_cluster_settings.
type tfModel struct {
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	Persistent              types.Object `tfsdk:"persistent"`
	Transient               types.Object `tfsdk:"transient"`
}

func (m tfModel) GetID() types.String                    { return m.ID }
func (m tfModel) GetResourceID() types.String            { return types.StringValue(resourceID) }
func (m tfModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }

// settingsBlockModel represents the persistent/transient block, which contains a set of settings.
type settingsBlockModel struct {
	Setting types.Set `tfsdk:"setting"`
}

// settingModel represents a single key-value setting entry.
type settingModel struct {
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	ValueList types.List   `tfsdk:"value_list"`
}
