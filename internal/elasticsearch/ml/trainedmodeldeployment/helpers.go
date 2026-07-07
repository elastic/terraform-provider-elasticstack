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
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

func toAdaptiveAllocationsSettings(aa *AdaptiveAllocationsData) *types.AdaptiveAllocationsSettings {
	if aa == nil || aa.Enabled.IsNull() {
		return nil
	}
	settings := &types.AdaptiveAllocationsSettings{
		Enabled: aa.Enabled.ValueBool(),
	}
	if !aa.MinNumberOfAllocations.IsNull() {
		settings.MinNumberOfAllocations = typeutils.OptionalInt(aa.MinNumberOfAllocations)
	}
	if !aa.MaxNumberOfAllocations.IsNull() {
		settings.MaxNumberOfAllocations = typeutils.OptionalInt(aa.MaxNumberOfAllocations)
	}
	return settings
}

func populateComputedFromStats(data *TrainedModelDeploymentData, stats *types.TrainedModelStats, statsJSON string) {
	if stats.DeploymentStats.State != nil {
		data.State = fwtypes.StringValue(stats.DeploymentStats.State.String())
	} else {
		data.State = fwtypes.StringNull()
	}
	if stats.DeploymentStats.AllocationStatus != nil {
		data.AllocationStatus = fwtypes.StringValue(stats.DeploymentStats.AllocationStatus.State.String())
	} else {
		data.AllocationStatus = fwtypes.StringNull()
	}
	data.StatsJSON = fwtypes.StringValue(statsJSON)

	data.NumberOfAllocations = typeutils.IntPointerToInt64Value(stats.DeploymentStats.NumberOfAllocations)
}
