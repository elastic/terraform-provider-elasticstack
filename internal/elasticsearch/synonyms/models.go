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

package synonyms

import (
	"context"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// synonymsAttrName is the tfsdk attribute name for synonym rule strings. Used as
// a constant to satisfy the goconst linter across schema and model definitions.
const synonymsAttrName = "synonyms"

// SynonymRuleModel represents a single synonym rule nested block.
type SynonymRuleModel struct {
	// ID is optional and computed: when omitted, the provider generates a UUID.
	ID       fwtypes.String `tfsdk:"id"`
	Synonyms fwtypes.String `tfsdk:"synonyms"`
}

// synonymRuleModelAttrTypes returns the attr.Type map for a SynonymRuleModel
// element, matching the tfsdk tags in SynonymRuleModel.
func synonymRuleModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             fwtypes.StringType,
		synonymsAttrName: fwtypes.StringType,
	}
}

// SynonymSetData is the Terraform state model for the elasticstack_elasticsearch_synonym_set resource.
// It implements entitycore.ElasticsearchResourceModel.
type SynonymSetData struct {
	entitycore.ResourceTimeoutsField
	entitycore.ElasticsearchConnectionField
	ID           fwtypes.String `tfsdk:"id"`
	SynonymSetID fwtypes.String `tfsdk:"synonym_set_id"`
	SynonymsSet  fwtypes.List   `tfsdk:"synonyms_set"`
}

func (data SynonymSetData) GetID() fwtypes.String         { return data.ID }
func (data SynonymSetData) GetResourceID() fwtypes.String { return data.SynonymSetID }
func (data SynonymSetData) GetElasticsearchConnection() fwtypes.List {
	return data.ElasticsearchConnection
}

// populateFromAPI maps []types.SynonymRuleRead (from the Elasticsearch API) into
// data.SynonymsSet as a types.List of SynonymRuleModel elements. Ordering is
// preserved as returned by the API.
func (data *SynonymSetData) populateFromAPI(ctx context.Context, rules []types.SynonymRuleRead, diagnostics *diag.Diagnostics) {
	models := make([]SynonymRuleModel, len(rules))
	for i, rule := range rules {
		models[i] = SynonymRuleModel{
			ID:       fwtypes.StringValue(rule.Id),
			Synonyms: fwtypes.StringValue(rule.Synonyms),
		}
	}

	list, d := fwtypes.ListValueFrom(ctx, fwtypes.ObjectType{AttrTypes: synonymRuleModelAttrTypes()}, models)
	diagnostics.Append(d...)
	if diagnostics.HasError() {
		return
	}

	data.SynonymsSet = list
}

// toAPIRules converts data.SynonymsSet into []types.SynonymRule for use in
// API calls. For each rule, if the ID is null, unknown, or empty a new UUID is
// generated; otherwise the stored ID value is used.
func (data SynonymSetData) toAPIRules(ctx context.Context, diagnostics *diag.Diagnostics) []types.SynonymRule {
	var models []SynonymRuleModel
	diagnostics.Append(data.SynonymsSet.ElementsAs(ctx, &models, false)...)
	if diagnostics.HasError() {
		return nil
	}

	rules := make([]types.SynonymRule, len(models))
	for i, model := range models {
		var ruleID string
		if !typeutils.IsKnown(model.ID) || model.ID.ValueString() == "" {
			ruleID = uuid.New().String()
		} else {
			ruleID = model.ID.ValueString()
		}
		rules[i] = types.SynonymRule{
			Id:       &ruleID,
			Synonyms: model.Synonyms.ValueString(),
		}
	}

	return rules
}
