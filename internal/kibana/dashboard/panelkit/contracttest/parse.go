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

package contracttest

import (
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// ParseDashboardPanel unmarshals JSON (one panel object from Kibana) into kbapi.DashboardPanelItem.
func ParseDashboardPanel(fullAPIResponse string) (kbapi.DashboardPanelItem, error) {
	var item kbapi.DashboardPanelItem
	if err := json.Unmarshal([]byte(fullAPIResponse), &item); err != nil {
		return kbapi.DashboardPanelItem{}, err
	}
	return item, nil
}
