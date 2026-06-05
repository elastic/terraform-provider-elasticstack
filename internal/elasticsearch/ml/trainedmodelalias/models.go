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

package trainedmodelalias

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TFModel struct {
	entitycore.ResourceTimeoutsField
	ID                      types.String `tfsdk:"id"`
	ElasticsearchConnection types.List   `tfsdk:"elasticsearch_connection"`
	ModelAlias              types.String `tfsdk:"model_alias"`
	ModelID                 types.String `tfsdk:"model_id"`
	Reassign                types.Bool   `tfsdk:"reassign"`
}

func (m TFModel) GetID() types.String { return m.ID }

func (m TFModel) GetResourceID() types.String { return m.ModelAlias }

func (m TFModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }
