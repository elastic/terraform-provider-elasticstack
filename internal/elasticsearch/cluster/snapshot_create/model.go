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

package snapshot_create

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	actiontimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model holds the Terraform configuration for the snapshot create action.
type Model struct {
	Repository              types.String         `tfsdk:"repository"`
	Snapshot                types.String         `tfsdk:"snapshot"`
	Indices                 types.List           `tfsdk:"indices"`
	IncludeGlobalState      types.Bool           `tfsdk:"include_global_state"`
	IgnoreUnavailable       types.Bool           `tfsdk:"ignore_unavailable"`
	Partial                 types.Bool           `tfsdk:"partial"`
	FeatureStates           types.List           `tfsdk:"feature_states"`
	ExpandWildcards         types.String         `tfsdk:"expand_wildcards"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	WaitForCompletion       types.Bool           `tfsdk:"wait_for_completion"`
	Timeouts                actiontimeouts.Value `tfsdk:"timeouts"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
}
