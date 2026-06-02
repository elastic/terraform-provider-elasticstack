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

package entities

import (
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	minVersion = version.Must(version.NewVersion("9.4.0"))
)

type dsModel struct {
	entitycore.KibanaConnectionField
	ID          types.String         `tfsdk:"id"`
	SpaceID     types.String         `tfsdk:"space_id"`
	EntityID    types.String         `tfsdk:"entity_id"`
	Filter      types.String         `tfsdk:"filter"`
	Size        types.Int64          `tfsdk:"size"`
	SearchAfter types.String         `tfsdk:"search_after"`
	Source      types.List           `tfsdk:"source"`
	Fields      types.List           `tfsdk:"fields"`
	SortField   types.String         `tfsdk:"sort_field"`
	SortOrder   types.String         `tfsdk:"sort_order"`
	Page        types.Int64          `tfsdk:"page"`
	PerPage     types.Int64          `tfsdk:"per_page"`
	FilterQuery types.String         `tfsdk:"filter_query"`
	EntityTypes types.Set            `tfsdk:"entity_types"`
	ResultsJSON jsontypes.Normalized `tfsdk:"results_json"`
	Items       types.List           `tfsdk:"items"`
}

var _ entitycore.WithVersionRequirements = (*dsModel)(nil)

func (*dsModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *minVersion,
		ErrorMessage: fmt.Sprintf("elasticstack_kibana_security_entity_store_entities is supported only for Kibana v%s and above", minVersion.String()),
	}}, nil
}
