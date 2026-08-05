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

import (
	"encoding/json"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultFilterSimpleLanguage = "kql"

// stringUnionFromAPI extracts a string represented by a generated union.
func stringUnionFromAPI(value json.Marshaler) string {
	b, err := value.MarshalJSON()
	if err != nil {
		return ""
	}

	var result string
	if err := json.Unmarshal(b, &result); err != nil {
		return ""
	}
	return result
}

// stringUnionToAPI builds a generated string union from its wire value.
func stringUnionToAPI[T any](value string) *T {
	var result T
	if err := json.Unmarshal([]byte(strconv.Quote(value)), &result); err != nil {
		return nil
	}
	return &result
}

// FilterSimpleFromAPI maps kbapi.KibanaHTTPAPIsFilterSimple into FilterSimpleModel.
func FilterSimpleFromAPI(m *models.FilterSimpleModel, apiQuery *kbapi.KibanaHTTPAPIsFilterSimple) {
	if apiQuery == nil {
		m.Expression = types.StringValue("")
		m.Language = types.StringValue(defaultFilterSimpleLanguage)
		return
	}
	m.Expression = types.StringValue(apiQuery.Expression)
	if apiQuery.Language == nil {
		m.Language = types.StringValue(defaultFilterSimpleLanguage)
		return
	}
	m.Language = types.StringValue(stringUnionFromAPI(apiQuery.Language))
}

// ConfigUsesESQL reports whether a config's query field indicates an ES|QL data source.
// A nil query means no filter is set, which is the ES|QL default state.
// A query with null expression and language is the explicit ES|QL sentinel.
func ConfigUsesESQL(query *models.FilterSimpleModel) bool {
	if query == nil {
		return true
	}
	return query.Expression.IsNull() && query.Language.IsNull()
}

// FilterSimpleToAPI maps FilterSimpleModel into kbapi.KibanaHTTPAPIsFilterSimple.
func FilterSimpleToAPI(m *models.FilterSimpleModel) *kbapi.KibanaHTTPAPIsFilterSimple {
	if m == nil {
		return nil
	}

	query := &kbapi.KibanaHTTPAPIsFilterSimple{
		Expression: m.Expression.ValueString(),
	}
	if typeutils.IsKnown(m.Language) {
		query.Language = stringUnionToAPI[kbapi.KibanaHTTPAPIsFilterSimple_Language](m.Language.ValueString())
	}
	return query
}
