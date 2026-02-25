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

type filterSimpleModel struct {
	Language types.String `tfsdk:"language"`
	Query    types.String `tfsdk:"query"`
}

func (m *filterSimpleModel) fromAPI(apiQuery kbapi.FilterSimpleSchema) {
	m.Query = types.StringValue(apiQuery.Query)
	if apiQuery.Language == nil {
		m.Language = types.StringValue("kuery")
		return
	}
	m.Language = typeutils.StringishPointerValue(apiQuery.Language)
}

func (m *filterSimpleModel) toAPI() kbapi.FilterSimpleSchema {
	if m == nil {
		return kbapi.FilterSimpleSchema{}
	}

	query := kbapi.FilterSimpleSchema{
		Query: m.Query.ValueString(),
	}
	if typeutils.IsKnown(m.Language) {
		lang := kbapi.FilterSimpleSchemaLanguage(m.Language.ValueString())
		query.Language = &lang
	}
	return query
}
