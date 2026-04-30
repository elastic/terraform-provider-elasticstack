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

package datafeed_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
)

// makeExpandWildcardsValue is a test helper that builds an ExpandWildcardsValue
// from a slice of string tokens.
func makeExpandWildcardsValue(t *testing.T, tokens ...string) datafeed.ExpandWildcardsValue {
	t.Helper()
	elems := make([]attr.Value, len(tokens))
	for i, tok := range tokens {
		elems[i] = types.StringValue(tok)
	}
	v, diags := datafeed.NewExpandWildcardsValue(elems)
	require.False(t, diags.HasError(), "unexpected error building ExpandWildcardsValue: %v", diags)
	return v
}

func TestExpandWildcardsSemanticEquals_AllEqualsConstituents(t *testing.T) {
	ctx := context.Background()

	all := makeExpandWildcardsValue(t, "all")
	constituents := makeExpandWildcardsValue(t, "open", "closed", "hidden")

	// ["all"] == ["open","closed","hidden"]
	eq, diags := all.SetSemanticEquals(ctx, constituents)
	require.False(t, diags.HasError())
	assert.True(t, eq, `["all"] should be semantically equal to ["open","closed","hidden"]`)

	// ["open","closed","hidden"] == ["all"]  (both directions)
	eq, diags = constituents.SetSemanticEquals(ctx, all)
	require.False(t, diags.HasError())
	assert.True(t, eq, `["open","closed","hidden"] should be semantically equal to ["all"]`)
}

func TestExpandWildcardsSemanticEquals_OrderInsensitive(t *testing.T) {
	ctx := context.Background()

	a := makeExpandWildcardsValue(t, "open", "closed", "hidden")
	b := makeExpandWildcardsValue(t, "hidden", "open", "closed")

	eq, diags := a.SetSemanticEquals(ctx, b)
	require.False(t, diags.HasError())
	assert.True(t, eq, `set order should not matter`)
}

func TestExpandWildcardsSemanticEquals_PartialExpansionNotEqual(t *testing.T) {
	ctx := context.Background()

	all := makeExpandWildcardsValue(t, "all")
	partial := makeExpandWildcardsValue(t, "open", "closed")

	eq, diags := all.SetSemanticEquals(ctx, partial)
	require.False(t, diags.HasError())
	assert.False(t, eq, `["all"] should NOT equal ["open","closed"] (missing "hidden")`)
}

func TestExpandWildcardsSemanticEquals_NoneEqualsSelf(t *testing.T) {
	ctx := context.Background()

	none1 := makeExpandWildcardsValue(t, "none")
	none2 := makeExpandWildcardsValue(t, "none")

	eq, diags := none1.SetSemanticEquals(ctx, none2)
	require.False(t, diags.HasError())
	assert.True(t, eq, `["none"] should equal ["none"]`)
}

func TestExpandWildcardsSemanticEquals_NoneNotEqualOpen(t *testing.T) {
	ctx := context.Background()

	none := makeExpandWildcardsValue(t, "none")
	open := makeExpandWildcardsValue(t, "open")

	eq, diags := none.SetSemanticEquals(ctx, open)
	require.False(t, diags.HasError())
	assert.False(t, eq, `["none"] should NOT equal ["open"]`)
}

func TestExpandWildcardsSemanticEquals_NullEqualsNull(t *testing.T) {
	ctx := context.Background()

	null1 := datafeed.NewExpandWildcardsNull()
	null2 := datafeed.NewExpandWildcardsNull()

	eq, diags := null1.SetSemanticEquals(ctx, null2)
	require.False(t, diags.HasError())
	assert.True(t, eq, `null should equal null`)
}

func TestExpandWildcardsSemanticEquals_UnknownEqualsUnknown(t *testing.T) {
	ctx := context.Background()

	unknown1 := datafeed.NewExpandWildcardsUnknown()
	unknown2 := datafeed.NewExpandWildcardsUnknown()

	eq, diags := unknown1.SetSemanticEquals(ctx, unknown2)
	require.False(t, diags.HasError())
	assert.True(t, eq, `unknown should equal unknown`)
}

func TestExpandWildcardsSemanticEquals_NullNotEqualNonNull(t *testing.T) {
	ctx := context.Background()

	null := datafeed.NewExpandWildcardsNull()
	known := makeExpandWildcardsValue(t, "open")

	// null != non-null
	eq, diags := null.SetSemanticEquals(ctx, known)
	require.False(t, diags.HasError())
	assert.False(t, eq, `null should NOT equal a known value`)

	// non-null != null
	eq, diags = known.SetSemanticEquals(ctx, null)
	require.False(t, diags.HasError())
	assert.False(t, eq, `a known value should NOT equal null`)
}

func TestExpandWildcardsSemanticEquals_UnknownNotEqualKnown(t *testing.T) {
	ctx := context.Background()

	unknown := datafeed.NewExpandWildcardsUnknown()
	known := makeExpandWildcardsValue(t, "open")

	// unknown != known
	eq, diags := unknown.SetSemanticEquals(ctx, known)
	require.False(t, diags.HasError())
	assert.False(t, eq, `unknown should NOT equal a known value`)

	// known != unknown
	eq, diags = known.SetSemanticEquals(ctx, unknown)
	require.False(t, diags.HasError())
	assert.False(t, eq, `a known value should NOT equal unknown`)
}
