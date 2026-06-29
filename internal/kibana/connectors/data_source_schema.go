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

package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/kbschema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func getDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Search for a connector by name, space id, and type. Note, that this data source will fail if more than one connector shares the same name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
			},
			"space_id": kbschema.DataSourceSpaceIDAttribute(),
			attrName: schema.StringAttribute{
				Description: "The name of the connector. While this name does not have to be unique, a distinctive name can help you identify a connector.",
				Required:    true,
			},
			attrConnectorTypeID: schema.StringAttribute{
				Description: "The ID of the connector type, e.g. `.index`.",
				Optional:    true,
				Computed:    true,
			},
			"connector_id": schema.StringAttribute{
				Description: "A UUID v1 or v4 randomly generated ID.",
				Computed:    true,
			},
			attrConfig: schema.StringAttribute{
				Description: "The configuration for the connector. Configuration properties vary depending on the connector type.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
			},
			attrIsDeprecated: schema.BoolAttribute{
				Description: "Indicates whether the connector type is deprecated.",
				Computed:    true,
			},
			attrIsMissingSecrets: schema.BoolAttribute{
				Description: "Indicates whether secrets are missing for the connector.",
				Computed:    true,
			},
			attrIsPreconfigured: schema.BoolAttribute{
				Description: "Indicates whether it is a preconfigured connector.",
				Computed:    true,
			},
		},
	}
}
