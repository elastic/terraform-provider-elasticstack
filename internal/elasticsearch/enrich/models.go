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

package enrich

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyData struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	PolicyType              types.String         `tfsdk:"policy_type"`
	Indices                 types.Set            `tfsdk:"indices"`
	MatchField              types.String         `tfsdk:"match_field"`
	EnrichFields            types.Set            `tfsdk:"enrich_fields"`
	Query                   jsontypes.Normalized `tfsdk:"query"`
}

type PolicyDataWithExecute struct {
	PolicyData
	Execute types.Bool `tfsdk:"execute"`
}

// populateFromPolicy converts models.EnrichPolicy to PolicyData fields
func (data *PolicyData) populateFromPolicy(ctx context.Context, policy *models.EnrichPolicy, diagnostics *diag.Diagnostics) {
	data.Name = types.StringValue(policy.Name)
	data.PolicyType = types.StringValue(policy.Type)
	data.MatchField = types.StringValue(policy.MatchField)

	if policy.Query != "" && policy.Query != "null" {
		data.Query = jsontypes.NewNormalizedValue(policy.Query)
	} else {
		data.Query = jsontypes.NewNormalizedNull()
	}

	// Convert string slices to Set
	data.Indices = typeutils.SetValueFrom(ctx, policy.Indices, types.StringType, path.Empty(), diagnostics)
	if diagnostics.HasError() {
		return
	}

	data.EnrichFields = typeutils.SetValueFrom(ctx, policy.EnrichFields, types.StringType, path.Empty(), diagnostics)
}
