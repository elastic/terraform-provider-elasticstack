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

package indexmappings

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID                      types.String         `tfsdk:"id"`
	Index                   types.String         `tfsdk:"index"`
	Mappings                index.MappingsValue  `tfsdk:"mappings"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
}

func (model tfModel) GetID() types.String                    { return model.ID }
func (model tfModel) GetResourceID() types.String            { return model.Index }
func (model tfModel) GetElasticsearchConnection() types.List { return model.ElasticsearchConnection }
