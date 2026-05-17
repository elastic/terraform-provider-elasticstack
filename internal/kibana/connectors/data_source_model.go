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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// connectorDataSourceModel maps the action connector data source schema.
type connectorDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID                types.String         `tfsdk:"id"`
	SpaceID           types.String         `tfsdk:"space_id"`
	Name              types.String         `tfsdk:"name"`
	ConnectorTypeID   types.String         `tfsdk:"connector_type_id"`
	ConnectorID       types.String         `tfsdk:"connector_id"`
	Config            jsontypes.Normalized `tfsdk:"config"`
	IsDeprecated      types.Bool           `tfsdk:"is_deprecated"`
	IsMissingSecrets  types.Bool           `tfsdk:"is_missing_secrets"`
	IsPreconfigured   types.Bool           `tfsdk:"is_preconfigured"`
}
