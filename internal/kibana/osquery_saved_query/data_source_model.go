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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ entitycore.WithVersionRequirements = dataSourceModel{}

type dataSourceModel struct {
	osquerySavedQueryBaseModel
	Prebuilt types.Bool `tfsdk:"prebuilt"`
}

func (m dataSourceModel) GetVersionRequirements(ctx context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return m.osquerySavedQueryBaseModel.GetVersionRequirements(ctx)
}

func (m *dataSourceModel) populateFromGetAPI(ctx context.Context, entity *kibanaoapi.OsquerySavedQueryGetEntity) diag.Diagnostics {
	if entity == nil {
		return nil
	}

	diags := m.osquerySavedQueryBaseModel.populateFromGetAPI(ctx, entity)
	if diags.HasError() {
		return diags
	}

	m.Prebuilt = prebuiltFromAPI(entity.Prebuilt)

	return diags
}

// prebuiltFromAPI maps the API prebuilt flag to state. Omitted/nil is treated as false
// so user-managed queries surface prebuilt = false rather than null.
func prebuiltFromAPI(prebuilt *bool) types.Bool {
	if prebuilt == nil {
		return types.BoolValue(false)
	}

	return types.BoolValue(*prebuilt)
}
