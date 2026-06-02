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

package trainedmodeldeployment

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data represents the Terraform state model for the trained model deployment resource.
//
//nolint:revive // consistent naming with other ML resources
type TrainedModelDeploymentData struct {
	ID                      types.String             `tfsdk:"id"`
	ElasticsearchConnection types.List               `tfsdk:"elasticsearch_connection"`
	ModelID                 types.String             `tfsdk:"model_id"`
	DeploymentID            types.String             `tfsdk:"deployment_id"`
	NumberOfAllocations     types.Int64              `tfsdk:"number_of_allocations"`
	ThreadsPerAllocation    types.Int64              `tfsdk:"threads_per_allocation"`
	Priority                types.String             `tfsdk:"priority"`
	QueueCapacity           types.Int64              `tfsdk:"queue_capacity"`
	WaitFor                 types.String             `tfsdk:"wait_for"`
	APITimeout              customtypes.Duration     `tfsdk:"api_timeout"`
	ForceStop               types.Bool               `tfsdk:"force_stop"`
	AdaptiveAllocations     *AdaptiveAllocationsData `tfsdk:"adaptive_allocations"`
	Timeouts                timeouts.Value           `tfsdk:"timeouts"`
	State                   types.String             `tfsdk:"state"`
	AllocationStatus        types.String             `tfsdk:"allocation_status"`
	StatsJSON               types.String             `tfsdk:"stats_json"`
}

type AdaptiveAllocationsData struct {
	Enabled                types.Bool  `tfsdk:"enabled"`
	MinNumberOfAllocations types.Int64 `tfsdk:"min_number_of_allocations"`
	MaxNumberOfAllocations types.Int64 `tfsdk:"max_number_of_allocations"`
}

func (d TrainedModelDeploymentData) GetID() types.String         { return d.ID }
func (d TrainedModelDeploymentData) GetResourceID() types.String { return d.DeploymentID }
func (d TrainedModelDeploymentData) GetElasticsearchConnection() types.List {
	return d.ElasticsearchConnection
}

var (
	_ entitycore.ElasticsearchResourceModel = TrainedModelDeploymentData{}
	_ entitycore.WithOptionalWriteIdentity  = TrainedModelDeploymentData{}
	_ entitycore.WithReadResourceID         = TrainedModelDeploymentData{}
)

// AllowsEmptyWriteIdentityOnCreate satisfies [entitycore.WithOptionalWriteIdentity].
func (TrainedModelDeploymentData) AllowsEmptyWriteIdentityOnCreate() bool { return true }

// GetReadResourceID satisfies [entitycore.WithReadResourceID].
func (d TrainedModelDeploymentData) GetReadResourceID() string {
	if typeutils.IsKnown(d.DeploymentID) {
		return d.DeploymentID.ValueString()
	}
	if typeutils.IsKnown(d.ModelID) {
		return d.ModelID.ValueString()
	}
	return ""
}
