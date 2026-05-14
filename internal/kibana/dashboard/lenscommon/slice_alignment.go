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

package lenscommon

import "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"

var sliceAligners []func(planPanels, statePanels []models.PanelModel)

// RegisterSliceAligner registers a function that aligns Terraform state against plan using the full panel slice.
// Aligners run in registration order when ApplySliceAligners is invoked.
func RegisterSliceAligner(f func(planPanels, statePanels []models.PanelModel)) {
	sliceAligners = append(sliceAligners, f)
}

// ApplySliceAligners runs all registered slice aligners in registration order.
func ApplySliceAligners(planPanels, statePanels []models.PanelModel) {
	for _, f := range sliceAligners {
		f(planPanels, statePanels)
	}
}
