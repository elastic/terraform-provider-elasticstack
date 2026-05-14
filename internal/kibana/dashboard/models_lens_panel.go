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

package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

// lensVisualizationBase is embedded by legacy per-chart panel converters until those structs are removed.
type lensVisualizationBase struct {
	visualizationType string
	hasTFChartBlock   func(blocks *models.LensByValueChartBlocks) bool
}

func (c lensVisualizationBase) visType() string {
	return c.visualizationType
}

func (c lensVisualizationBase) handlesTFConfigBlocks(blocks *models.LensByValueChartBlocks) bool {
	if blocks == nil || c.hasTFChartBlock == nil {
		return false
	}
	return c.hasTFChartBlock(blocks)
}
