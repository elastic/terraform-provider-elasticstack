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

package typeutils_test

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestStringTypableValueFromTerraform(t *testing.T) {
	t.Run("returns an error if the tf value is not a string", func(t *testing.T) {
		val, err := typeutils.StringTypableValueFromTerraform(
			context.Background(),
			basetypes.StringType{},
			func(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
				return in, nil
			},
			tftypes.NewValue(tftypes.Bool, true),
		)

		require.Nil(t, val)
		require.ErrorContains(t, err, "expected string")
	})

	t.Run("returns an error if valueFromString reports a diagnostic error", func(t *testing.T) {
		val, err := typeutils.StringTypableValueFromTerraform(
			context.Background(),
			basetypes.StringType{},
			func(_ context.Context, _ basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
				var diags diag.Diagnostics
				diags.AddError("boom", "boom")
				return nil, diags
			},
			tftypes.NewValue(tftypes.String, "hello"),
		)

		require.Nil(t, val)
		require.ErrorContains(t, err, "unexpected error converting StringValue to StringValuable")
	})

	t.Run("delegates to valueFromString on success", func(t *testing.T) {
		val, err := typeutils.StringTypableValueFromTerraform(
			context.Background(),
			basetypes.StringType{},
			func(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
				return in, nil
			},
			tftypes.NewValue(tftypes.String, "hello"),
		)

		require.NoError(t, err)
		require.Equal(t, basetypes.NewStringValue("hello"), val)
	})
}

type fakeStringType struct {
	basetypes.StringType
}

type otherStringType struct {
	basetypes.StringType
}

func TestStringTypableEqual(t *testing.T) {
	t.Run("true when the other type is the same concrete type", func(t *testing.T) {
		require.True(t, typeutils.StringTypableEqual[fakeStringType](fakeStringType{}))
	})

	t.Run("false when the other type is a different concrete type", func(t *testing.T) {
		require.False(t, typeutils.StringTypableEqual[fakeStringType](otherStringType{}))
	})

	t.Run("false when the other type is not an attr.Type StringTypable at all", func(t *testing.T) {
		require.False(t, typeutils.StringTypableEqual[fakeStringType](basetypes.BoolType{}))
	})
}

type fakeStringValue struct {
	basetypes.StringValue
}

func fakeStringValueOf(v fakeStringValue) basetypes.StringValue {
	return v.StringValue
}

func TestStringValuableEqual(t *testing.T) {
	t.Run("false when the other value is not the same concrete type", func(t *testing.T) {
		self := fakeStringValue{StringValue: basetypes.NewStringValue("hello")}
		require.False(t, typeutils.StringValuableEqual(self.StringValue, basetypes.NewStringValue("hello"), fakeStringValueOf))
	})

	t.Run("false when the embedded values differ", func(t *testing.T) {
		self := fakeStringValue{StringValue: basetypes.NewStringValue("hello")}
		other := fakeStringValue{StringValue: basetypes.NewStringValue("world")}
		require.False(t, typeutils.StringValuableEqual(self.StringValue, other, fakeStringValueOf))
	})

	t.Run("true when the embedded values are equal", func(t *testing.T) {
		self := fakeStringValue{StringValue: basetypes.NewStringValue("hello")}
		other := fakeStringValue{StringValue: basetypes.NewStringValue("hello")}
		require.True(t, typeutils.StringValuableEqual(self.StringValue, other, fakeStringValueOf))
	})
}
