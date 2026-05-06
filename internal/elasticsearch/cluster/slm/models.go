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

package slm

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data holds the Terraform state for the snapshot lifecycle resource.
type Data struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Schedule                types.String         `tfsdk:"schedule"`
	Repository              types.String         `tfsdk:"repository"`
	SnapshotName            types.String         `tfsdk:"snapshot_name"`
	ExpandWildcards         types.String         `tfsdk:"expand_wildcards"`
	IgnoreUnavailable       types.Bool           `tfsdk:"ignore_unavailable"`
	IncludeGlobalState      types.Bool           `tfsdk:"include_global_state"`
	Indices                 types.List           `tfsdk:"indices"`
	FeatureStates           types.Set            `tfsdk:"feature_states"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Partial                 types.Bool           `tfsdk:"partial"`
	ExpireAfter             types.String         `tfsdk:"expire_after"`
	MaxCount                types.Int64          `tfsdk:"max_count"`
	MinCount                types.Int64          `tfsdk:"min_count"`
}

func (d Data) GetID() types.String                    { return d.ID }
func (d Data) GetResourceID() types.String            { return d.Name }
func (d Data) GetElasticsearchConnection() types.List { return d.ElasticsearchConnection }
