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

package enrollmenttokens

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type enrollmentTokensModel struct {
	ID       types.String `tfsdk:"id"`
	PolicyID types.String `tfsdk:"policy_id"`
	SpaceID  types.String `tfsdk:"space_id"`
	Tokens   types.List   `tfsdk:"tokens"` // > enrollmentTokenModel
}

type enrollmentTokenModel struct {
	KeyID     types.String `tfsdk:"key_id"`
	APIKey    types.String `tfsdk:"api_key"`
	APIKeyID  types.String `tfsdk:"api_key_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Name      types.String `tfsdk:"name"`
	Active    types.Bool   `tfsdk:"active"`
	PolicyID  types.String `tfsdk:"policy_id"`
}

func (model *enrollmentTokensModel) populateFromAPI(ctx context.Context, data []kbapi.EnrollmentApiKey) (diags diag.Diagnostics) {
	model.Tokens = typeutils.SliceToListType(ctx, data, getTokenType(), path.Root("tokens"), &diags, newEnrollmentTokenModel)
	return
}

func newEnrollmentTokenModel(data kbapi.EnrollmentApiKey, _ typeutils.ListMeta) enrollmentTokenModel {
	return enrollmentTokenModel{
		KeyID:     types.StringValue(data.Id),
		Active:    types.BoolValue(data.Active),
		APIKey:    types.StringValue(data.ApiKey),
		APIKeyID:  types.StringValue(data.ApiKeyId),
		CreatedAt: types.StringValue(data.CreatedAt),
		Name:      types.StringPointerValue(data.Name),
		PolicyID:  types.StringPointerValue(data.PolicyId),
	}
}
