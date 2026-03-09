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

package streams

import (
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// queryConfigModel is the Terraform model for a query stream's configuration.
type queryConfigModel struct {
	Esql types.String `tfsdk:"esql"`
	View types.String `tfsdk:"view"`
}

// populateFromAPI populates the query config model from an API response.
func (m *queryConfigModel) populateFromAPI(q *kibanaoapi.StreamQueryESQLDef) {
	if q == nil {
		return
	}
	m.Esql = types.StringValue(q.Esql)
	if q.View != "" {
		m.View = types.StringValue(q.View)
	} else {
		m.View = types.StringNull()
	}
}

// toAPI converts the query config model to an API query definition.
func (m *queryConfigModel) toAPI() *kibanaoapi.StreamQueryESQLDef {
	q := &kibanaoapi.StreamQueryESQLDef{
		Esql: m.Esql.ValueString(),
	}
	if !m.View.IsNull() && !m.View.IsUnknown() {
		q.View = m.View.ValueString()
	}
	return q
}
