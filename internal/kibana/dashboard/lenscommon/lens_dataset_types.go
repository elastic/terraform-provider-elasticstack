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

import "encoding/json"

// LensDatasetTypeESQL and LensDatasetTypeTable are JSON type discriminators for Lens ES|QL/table datasets.
const (
	LensDatasetTypeESQL  = "esql"
	LensDatasetTypeTable = "table"
)

// LensDataSourceIsESQLOrTable reports whether a Lens chart data_source union JSON is ES|QL or table shaped.
func LensDataSourceIsESQLOrTable(body []byte, err error) bool {
	if err != nil {
		return false
	}
	var ds struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &ds); err != nil {
		return false
	}
	return ds.Type == LensDatasetTypeESQL || ds.Type == LensDatasetTypeTable
}
