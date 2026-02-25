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
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccessControlValue maps to the access_control block
type AccessControlValue struct {
	AccessMode types.String `tfsdk:"access_mode"`
	Owner      types.String `tfsdk:"owner"`
}

type accessControlAPIPostModel = struct {
	AccessMode *kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode `json:"access_mode,omitempty"`
	Owner      *string                                                  `json:"owner,omitempty"`
}

type accessControlAPIPutModel = struct {
	AccessMode *kbapi.PutDashboardsIdJSONBodyDataAccessControlAccessMode `json:"access_mode,omitempty"`
	Owner      *string                                                   `json:"owner,omitempty"`
}

// ToCreateAPI converts the Terraform model to the POST API model
func (m *AccessControlValue) toCreateAPI() *accessControlAPIPostModel {
	if m == nil {
		return nil
	}

	apiModel := &accessControlAPIPostModel{}

	if typeutils.IsKnown(m.AccessMode) {
		apiModel.AccessMode = new(kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode(m.AccessMode.ValueString()))
	}

	if typeutils.IsKnown(m.Owner) {
		apiModel.Owner = new(m.Owner.ValueString())
	}

	return apiModel
}

// ToUpdateAPI converts the Terraform model to the PUT API model
func (m *AccessControlValue) toUpdateAPI() *accessControlAPIPutModel {
	createModel := m.toCreateAPI()
	if createModel == nil {
		return nil
	}

	return &accessControlAPIPutModel{
		AccessMode: (*kbapi.PutDashboardsIdJSONBodyDataAccessControlAccessMode)(createModel.AccessMode),
		Owner:      createModel.Owner,
	}
}

// newAccessControlFromAPI maps the API response to the Terraform model
func newAccessControlFromAPI(accessMode *string, owner *string) *AccessControlValue {
	if accessMode == nil && owner == nil {
		return nil
	}

	return &AccessControlValue{
		AccessMode: types.StringPointerValue(accessMode),
		Owner:      types.StringPointerValue(owner),
	}
}
