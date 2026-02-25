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

package connectors

import (
	"context"
	"fmt"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*ConfigType)(nil)
)

type ConfigType struct {
	customtypes.JSONWithContextualDefaultsType
}

func NewConfigType() ConfigType {
	return ConfigType{
		JSONWithContextualDefaultsType: customtypes.NewJSONWithContextualDefaultsType(kibanaoapi.ConnectorConfigWithDefaults),
	}
}

// String returns a human readable string of the type name.
func (t ConfigType) String() string {
	return "connectors.ConfigType"
}

// ValueType returns the Value type.
func (t ConfigType) ValueType(ctx context.Context) attr.Value {
	return ConfigValue{
		JSONWithContextualDefaultsValue: t.JSONWithContextualDefaultsType.ValueType(ctx).(customtypes.JSONWithContextualDefaultsValue),
	}
}

// Equal returns true if the given type is equivalent.
func (t ConfigType) Equal(o attr.Type) bool {
	other, ok := o.(ConfigType)

	if !ok {
		return false
	}

	return t.JSONWithContextualDefaultsType.Equal(other.JSONWithContextualDefaultsType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t ConfigType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	val, diags := t.JSONWithContextualDefaultsType.ValueFromString(ctx, in)
	if diags.HasError() {
		return nil, diags
	}

	return ConfigValue{
		JSONWithContextualDefaultsValue: val.(customtypes.JSONWithContextualDefaultsValue),
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t ConfigType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
