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

package kbapi

// Hand-maintained companion for generated PutDashboardsIdJSONBody_Filters_Item.
//
// The kibana.gen.go type uses an unexported json.RawMessage union but omits the MarshalJSON/UnmarshalJSON
// hooks that exist on KbnDashboardData_Filters_Item. Without these methods, decoding filter JSON into the
// PUT request body leaves an empty union and updates cannot serialize filters.
//
// Removal: delete this file if codegen ever adds the same methods to PutDashboardsIdJSONBody_Filters_Item
// (duplicate receiver definitions will not compile).

func (t PutDashboardsIdJSONBody_Filters_Item) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

func (t *PutDashboardsIdJSONBody_Filters_Item) UnmarshalJSON(b []byte) error {
	err := t.union.UnmarshalJSON(b)
	return err
}
