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

package settings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingNameUniqueValidator_Duplicate_Error(t *testing.T) {
	ctx := context.Background()
	v := settingNameUniqueValidator{}

	obj1, diags := types.ObjectValue(settingModelAttrTypes(), map[string]attr.Value{
		"name":       types.StringValue("dup"),
		"value":      types.StringValue("v1"),
		"value_list": types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	obj2, diags := types.ObjectValue(settingModelAttrTypes(), map[string]attr.Value{
		"name":       types.StringValue("dup"),
		"value":      types.StringValue("v2"),
		"value_list": types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	setVal, diags := types.SetValue(types.ObjectType{AttrTypes: settingModelAttrTypes()}, []attr.Value{obj1, obj2})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	resp := &validator.SetResponse{}
	v.ValidateSet(ctx, validator.SetRequest{
		ConfigValue: setVal,
		Path:        path.Root("setting"),
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for duplicate setting name")
	}
}

func TestSettingNameUniqueValidator_Unique_OK(t *testing.T) {
	ctx := context.Background()
	v := settingNameUniqueValidator{}

	obj1, diags := types.ObjectValue(settingModelAttrTypes(), map[string]attr.Value{
		"name":       types.StringValue("a"),
		"value":      types.StringValue("v1"),
		"value_list": types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	obj2, diags := types.ObjectValue(settingModelAttrTypes(), map[string]attr.Value{
		"name":       types.StringValue("b"),
		"value":      types.StringValue("v2"),
		"value_list": types.ListNull(types.StringType),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	setVal, diags := types.SetValue(types.ObjectType{AttrTypes: settingModelAttrTypes()}, []attr.Value{obj1, obj2})
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	resp := &validator.SetResponse{}
	v.ValidateSet(ctx, validator.SetRequest{
		ConfigValue: setVal,
		Path:        path.Root("setting"),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
}
