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

package snapshot_restore

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model holds the Terraform configuration for the snapshot restore action.
// The elasticsearch_connection and timeouts blocks are provided by the
// embedded envelope fields and injected into the schema by
// [entitycore.NewElasticsearchAction].
type Model struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ActionTimeoutsField

	Repository          types.String         `tfsdk:"repository"`
	Snapshot            types.String         `tfsdk:"snapshot"`
	Indices             types.List           `tfsdk:"indices"`
	IncludeGlobalState  types.Bool           `tfsdk:"include_global_state"`
	IgnoreUnavailable   types.Bool           `tfsdk:"ignore_unavailable"`
	IncludeAliases      types.Bool           `tfsdk:"include_aliases"`
	Partial             types.Bool           `tfsdk:"partial"`
	FeatureStates       types.List           `tfsdk:"feature_states"`
	RenamePattern       types.String         `tfsdk:"rename_pattern"`
	RenameReplacement   types.String         `tfsdk:"rename_replacement"`
	IgnoreIndexSettings types.List           `tfsdk:"ignore_index_settings"`
	IndexSettings       jsontypes.Normalized `tfsdk:"index_settings"`
	WaitForCompletion   types.Bool           `tfsdk:"wait_for_completion"`
}
