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

package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// GridFromAPI maps API grid coords into Terraform panel grid fields.
func GridFromAPI(x, y float32, w, h *float32) models.PanelGridModel {
	g := models.PanelGridModel{
		X: types.Int64Value(int64(x)),
		Y: types.Int64Value(int64(y)),
	}
	if w != nil {
		g.W = types.Int64Value(int64(*w))
	} else {
		g.W = types.Int64Null()
	}
	if h != nil {
		g.H = types.Int64Value(int64(*h))
	} else {
		g.H = types.Int64Null()
	}
	return g
}

// GridToAPI converts Terraform grid into the serialized API grid shape passed to kbapi structs.
func GridToAPI(g models.PanelGridModel) struct {
	H *float32 `json:"h,omitempty"`
	W *float32 `json:"w,omitempty"`
	X float32  `json:"x"`
	Y float32  `json:"y"`
} {
	grid := struct {
		H *float32 `json:"h,omitempty"`
		W *float32 `json:"w,omitempty"`
		X float32  `json:"x"`
		Y float32  `json:"y"`
	}{
		X: float32(g.X.ValueInt64()),
		Y: float32(g.Y.ValueInt64()),
	}
	if typeutils.IsKnown(g.W) {
		w := float32(g.W.ValueInt64())
		grid.W = &w
	}
	if typeutils.IsKnown(g.H) {
		h := float32(g.H.ValueInt64())
		grid.H = &h
	}
	return grid
}

// IDFromAPI maps an optional API panel id into Terraform state.
func IDFromAPI(id *string) types.String {
	return types.StringPointerValue(id)
}

// IDToAPI maps Terraform state into an optional API panel id pointer.
func IDToAPI(id types.String) *string {
	if !typeutils.IsKnown(id) {
		return nil
	}
	s := id.ValueString()
	return &s
}
