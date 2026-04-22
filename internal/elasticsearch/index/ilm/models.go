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

package ilm

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tfModel struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	Name                    types.String         `tfsdk:"name"`
	Metadata                jsontypes.Normalized `tfsdk:"metadata"`
	Hot                     types.Object         `tfsdk:"hot"`
	Warm                    types.Object         `tfsdk:"warm"`
	Cold                    types.Object         `tfsdk:"cold"`
	Frozen                  types.Object         `tfsdk:"frozen"`
	Delete                  types.Object         `tfsdk:"delete"`
	ModifiedDate            types.String         `tfsdk:"modified_date"`
}

func (m *tfModel) phaseObject(name string) types.Object {
	switch name {
	case ilmPhaseHot:
		return m.Hot
	case ilmPhaseWarm:
		return m.Warm
	case ilmPhaseCold:
		return m.Cold
	case ilmPhaseFrozen:
		return m.Frozen
	case ilmPhaseDelete:
		return m.Delete
	default:
		return phaseObjectNull(ilmPhaseHot)
	}
}
