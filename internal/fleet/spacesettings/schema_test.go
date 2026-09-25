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

package spacesettings

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func validatePrefixes(t *testing.T, prefixes ...string) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()

	elements := make([]attr.Value, len(prefixes))
	for i, prefix := range prefixes {
		elements[i] = types.StringValue(prefix)
	}
	value := types.SetValueMust(types.StringType, elements)

	attribute, ok := getSchema(ctx).Attributes["allowed_namespace_prefixes"].(schema.SetAttribute)
	require.True(t, ok)

	var diags diag.Diagnostics
	for _, v := range attribute.Validators {
		resp := &validator.SetResponse{}
		v.ValidateSet(ctx, validator.SetRequest{Path: path.Root("allowed_namespace_prefixes"), ConfigValue: value}, resp)
		diags.Append(resp.Diagnostics...)
	}
	return diags
}

func TestSchema_rejectsMoreThanTenPrefixes(t *testing.T) {
	t.Parallel()

	prefixes := make([]string, 11)
	for i := range prefixes {
		prefixes[i] = fmt.Sprintf("prefix_%d", i)
	}

	require.True(t, validatePrefixes(t, prefixes...).HasError())
}

func TestSchema_acceptsEmptyPrefixList(t *testing.T) {
	t.Parallel()

	require.False(t, validatePrefixes(t).HasError())
}

func TestSchema_acceptsTenUniquePrefixes(t *testing.T) {
	t.Parallel()

	prefixes := make([]string, 10)
	for i := range prefixes {
		prefixes[i] = fmt.Sprintf("prefix_%d", i)
	}

	require.False(t, validatePrefixes(t, prefixes...).HasError())
}

func TestSchema_spaceIDIsRequiredAndForcesReplacement(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	spaceID, ok := getSchema(ctx).Attributes["space_id"].(schema.StringAttribute)
	require.True(t, ok)
	require.True(t, spaceID.Required)

	wantDescription := stringplanmodifier.RequiresReplace().Description(ctx)
	var descriptions []string
	for _, m := range spaceID.PlanModifiers {
		descriptions = append(descriptions, m.Description(ctx))
	}
	require.Contains(t, descriptions, wantDescription)
}

func TestSchema_idAndManagedByAreComputed(t *testing.T) {
	t.Parallel()
	attributes := getSchema(context.Background()).Attributes

	for _, name := range []string{"id", "managed_by"} {
		require.True(t, attributes[name].IsComputed(), name)
		require.False(t, attributes[name].IsRequired(), name)
	}
}
