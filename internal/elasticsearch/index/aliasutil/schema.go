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

package aliasutil

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// AliasSetNestedBlock returns the shared SetNestedBlock schema for alias blocks
// used in both component template and index template resources.
func AliasSetNestedBlock() schema.SetNestedBlock {
	return schema.SetNestedBlock{
		MarkdownDescription: "Alias to add.",
		NestedObject: schema.NestedBlockObject{
			CustomType: NewAliasObjectType(),
			Attributes: map[string]schema.Attribute{
				attrName: schema.StringAttribute{
					MarkdownDescription: "The alias name. Index alias names support date math. See the " +
						"[date math index names documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/date-math-index-names.html) " +
						"for more details.",
					Required: true,
				},
				attrFilter: schema.StringAttribute{
					MarkdownDescription: "Query used to limit documents the alias can access.",
					Optional:            true,
					CustomType:          jsontypes.NormalizedType{},
					Validators: []validator.String{
						validators.StringIsJSONObject{},
					},
				},
				attrIndexRouting: schema.StringAttribute{
					MarkdownDescription: "Value used to route indexing operations to a specific shard. If specified, this overwrites the routing value for indexing operations.",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
				attrIsHidden: schema.BoolAttribute{
					MarkdownDescription: "If true, the alias is hidden.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				attrIsWriteIndex: schema.BoolAttribute{
					MarkdownDescription: "If true, the index is the write index for the alias.",
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
				},
				attrRouting: schema.StringAttribute{
					MarkdownDescription: "Value used to route indexing and search operations to a specific shard.",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
				attrSearchRouting: schema.StringAttribute{
					MarkdownDescription: "Value used to route search operations to a specific shard. If specified, this overwrites the routing value for search operations.",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString(""),
				},
			},
		},
	}
}
