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

package securitylistdatastreams

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model represents the Terraform state/config model for the
// kibana_security_list_data_streams resource. This resource manages the creation of
// .lists and .items data streams required for security lists and exceptions.
type Model struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ListIndex     types.Bool   `tfsdk:"list_index"`
	ListItemIndex types.Bool   `tfsdk:"list_item_index"`
}

// fromAPIResponse populates the model from API response data.
// This helper method ensures consistency in how API responses are mapped to Terraform state.
func (m *Model) fromAPIResponse(spaceID string, listIndex, listItemIndex bool) {
	m.ID = types.StringValue(spaceID)
	m.SpaceID = types.StringValue(spaceID)
	m.ListIndex = types.BoolValue(listIndex)
	m.ListItemIndex = types.BoolValue(listItemIndex)
}
