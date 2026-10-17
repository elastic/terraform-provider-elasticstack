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

package agentdownloadsource

import (
	"context"
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_authCredentialsValidator(t *testing.T) {
	ctx := context.Background()
	authAttr, ok := getSchema(ctx).Attributes["auth"].(schema.SingleNestedAttribute)
	require.True(t, ok)
	attrTypes := authAttr.GetType().(types.ObjectType).AttrTypes

	str := types.StringValue
	null := types.StringNull()
	unknown := types.StringUnknown()

	build := func(overrides map[string]attr.Value) types.Object {
		vals := map[string]attr.Value{
			"username":    null,
			"password":    null,
			"password_wo": null,
			"api_key":     null,
			"api_key_wo":  null,
			"headers":     types.ListNull(attrTypes["headers"].(types.ListType).ElemType),
		}
		maps.Copy(vals, overrides)
		return types.ObjectValueMust(attrTypes, vals)
	}

	tests := []struct {
		name    string
		obj     types.Object
		wantErr bool
	}{
		{name: "null object", obj: types.ObjectNull(attrTypes)},
		{name: "unknown object defers", obj: types.ObjectUnknown(attrTypes)},
		{name: "empty auth is accepted", obj: build(nil)},
		{name: "username with password", obj: build(map[string]attr.Value{"username": str("u"), "password": str("p")})},
		{name: "username with password_wo", obj: build(map[string]attr.Value{"username": str("u"), "password_wo": str("p")})},
		{name: "username alone rejected", obj: build(map[string]attr.Value{"username": str("u")}), wantErr: true},
		{name: "password alone rejected", obj: build(map[string]attr.Value{"password": str("p")}), wantErr: true},
		{name: "password_wo alone rejected", obj: build(map[string]attr.Value{"password_wo": str("p")}), wantErr: true},
		{name: "password and password_wo rejected", obj: build(map[string]attr.Value{"username": str("u"), "password": str("p"), "password_wo": str("p")}), wantErr: true},
		{name: "api_key alone", obj: build(map[string]attr.Value{"api_key": str("k")})},
		{name: "api_key_wo alone", obj: build(map[string]attr.Value{"api_key_wo": str("k")})},
		{name: "api_key and api_key_wo rejected", obj: build(map[string]attr.Value{"api_key": str("k"), "api_key_wo": str("k")}), wantErr: true},
		{name: "password with api_key rejected", obj: build(map[string]attr.Value{"username": str("u"), "password": str("p"), "api_key": str("k")}), wantErr: true},
		{name: "password_wo with api_key rejected", obj: build(map[string]attr.Value{"username": str("u"), "password_wo": str("p"), "api_key": str("k")}), wantErr: true},
		{name: "password with api_key_wo rejected", obj: build(map[string]attr.Value{"username": str("u"), "password": str("p"), "api_key_wo": str("k")}), wantErr: true},
		{name: "password_wo with api_key_wo rejected", obj: build(map[string]attr.Value{"username": str("u"), "password_wo": str("p"), "api_key_wo": str("k")}), wantErr: true},
		{name: "unknown password defers pairing", obj: build(map[string]attr.Value{"username": str("u"), "password": unknown})},
		{name: "unknown username defers pairing", obj: build(map[string]attr.Value{"username": unknown, "password_wo": str("p")})},
		{name: "unknown api_key does not conflict", obj: build(map[string]attr.Value{"username": str("u"), "password": str("p"), "api_key": unknown})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp validator.ObjectResponse
			authCredentialsValidator{}.ValidateObject(ctx, validator.ObjectRequest{
				ConfigValue: tt.obj,
				Path:        path.Root("auth"),
			}, &resp)
			require.Equal(t, tt.wantErr, resp.Diagnostics.HasError(), resp.Diagnostics)
		})
	}
}
