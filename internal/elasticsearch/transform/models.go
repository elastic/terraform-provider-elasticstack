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

package transform

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tfModel is the Plugin Framework model for the transform resource.
type tfModel struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Description             types.String         `tfsdk:"description"`
	Source                  *tfModelSource       `tfsdk:"source"`
	Destination             *tfModelDestination  `tfsdk:"destination"`
	Pivot                   jsontypes.Normalized `tfsdk:"pivot"`
	Latest                  jsontypes.Normalized `tfsdk:"latest"`
	Frequency               types.String         `tfsdk:"frequency"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	RetentionPolicy         *tfModelRetention    `tfsdk:"retention_policy"`
	Sync                    *tfModelSync         `tfsdk:"sync"`
	AlignCheckpoints        types.Bool           `tfsdk:"align_checkpoints"`
	DatesAsEpochMillis      types.Bool           `tfsdk:"dates_as_epoch_millis"`
	DeduceMappings          types.Bool           `tfsdk:"deduce_mappings"`
	DocsPerSecond           types.Float64        `tfsdk:"docs_per_second"`
	MaxPageSearchSize       types.Int64          `tfsdk:"max_page_search_size"`
	NumFailureRetries       types.Int64          `tfsdk:"num_failure_retries"`
	Unattended              types.Bool           `tfsdk:"unattended"`
	DeferValidation         types.Bool           `tfsdk:"defer_validation"`
	Timeout                 customtypes.Duration `tfsdk:"timeout"`
	Enabled                 types.Bool           `tfsdk:"enabled"`
}

func (m tfModel) GetID() types.String                    { return m.ID }
func (m tfModel) GetResourceID() types.String            { return m.Name }
func (m tfModel) GetElasticsearchConnection() types.List { return m.ElasticsearchConnection }

// tfModelSource holds the transform source block.
type tfModelSource struct {
	Indices         []types.String       `tfsdk:"indices"`
	Query           jsontypes.Normalized `tfsdk:"query"`
	RuntimeMappings jsontypes.Normalized `tfsdk:"runtime_mappings"`
}

// tfModelAlias holds a single destination alias.
type tfModelAlias struct {
	Alias          types.String `tfsdk:"alias"`
	MoveOnCreation types.Bool   `tfsdk:"move_on_creation"`
}

// tfModelDestination holds the transform destination block.
type tfModelDestination struct {
	Index    types.String   `tfsdk:"index"`
	Aliases  []tfModelAlias `tfsdk:"aliases"`
	Pipeline types.String   `tfsdk:"pipeline"`
}

// tfModelRetentionTime holds the time sub-block of retention_policy.
type tfModelRetentionTime struct {
	Field  types.String `tfsdk:"field"`
	MaxAge types.String `tfsdk:"max_age"`
}

// tfModelRetention holds the retention_policy block.
type tfModelRetention struct {
	Time *tfModelRetentionTime `tfsdk:"time"`
}

// tfModelSyncTime holds the time sub-block of sync.
type tfModelSyncTime struct {
	Field types.String `tfsdk:"field"`
	Delay types.String `tfsdk:"delay"`
}

// tfModelSync holds the sync block.
type tfModelSync struct {
	Time *tfModelSyncTime `tfsdk:"time"`
}
