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

package main

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func structOf(fields []*types.Var, tags []string) *types.Struct {
	return types.NewStruct(fields, tags)
}

func field(name string) *types.Var {
	return types.NewField(token.NoPos, nil, name, types.Typ[types.String], false)
}

func TestJSONFieldsHonoursTagsHiddenFieldsAndDefaults(t *testing.T) {
	st := structOf(
		[]*types.Var{field("Tagged"), field("Hidden"), field("Plain"), field("private"), field("Form")},
		[]string{`json:"tagged,omitempty"`, `json:"-"`, ``, `json:"private"`, `form:"form_name"`},
	)

	fields := jsonFields(st)

	assert.Contains(t, fields, "tagged")
	assert.Contains(t, fields, "Plain")
	assert.Contains(t, fields, "form_name")
	assert.NotContains(t, fields, "Hidden")
	assert.NotContains(t, fields, "private")
}

func TestJSONFieldsFlattensEmbeddedStructsAndShallowestNameWins(t *testing.T) {
	inner := structOf([]*types.Var{field("Name"), field("Deep")}, []string{`json:"name"`, `json:"deep"`})
	embedded := types.NewField(token.NoPos, nil, "Inner", inner, true)
	outer := structOf([]*types.Var{embedded, field("Name")}, []string{``, `json:"name"`})

	fields := jsonFields(outer)

	assert.Contains(t, fields, "deep")
	assert.Same(t, inner, fields["deep"].owner)
	assert.Same(t, outer, fields["name"].owner)
}
