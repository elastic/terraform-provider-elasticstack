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
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// getSchemaFactory returns the schema for the index mappings resource without the
// elasticsearch_connection block; the envelope injects that block automatically.
func getSchemaFactory(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manage a user-declared subset of index mappings on an existing Elasticsearch index. Destroy is a no-op — field mappings are not removed.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID in the form `<cluster_uuid>/<index_name>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"index": schema.StringAttribute{
				Description: "Name of the target Elasticsearch index.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mappings": schema.StringAttribute{
				Description: "JSON mappings object to manage on the index. All top-level keys (`properties`, `dynamic`, `_source`, " +
					"`dynamic_templates`, `runtime`, etc.) are supported. Only the keys and fields declared here are tracked; " +
					"dynamic extras added by Elasticsearch are ignored. Destroying this resource does not remove mappings from the index (a no-op).",
				Required:   true,
				CustomType: index.MappingsType{},
				Validators: []validator.String{
					index.StringIsJSONObject{NonEmpty: true},
				},
			},
		},
	}
}
