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

package schema

import (
	"context"
	"testing"

	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchConnectionNullList_objectMatchesGetEsFWConnectionBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	block := GetEsFWConnectionBlock()
	lb, ok := block.(fwschema.ListNestedBlock)
	if !ok {
		t.Fatal("GetEsFWConnectionBlock must return a ListNestedBlock")
	}
	want, err := fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
	if err != nil {
		t.Fatalf("fwNestedBlockAttributesToAttrTypes: %v", err)
	}

	list := ElasticsearchConnectionNullList()
	listType, ok := list.Type(ctx).(types.ListType)
	if !ok {
		t.Fatalf("ElasticsearchConnectionNullList.Type: got %T", list.Type(ctx))
	}
	objType, ok := listType.ElementType().(types.ObjectType)
	if !ok {
		t.Fatalf("list element type: got %T", listType.ElementType())
	}
	got := objType.AttrTypes

	if len(got) != len(want) {
		t.Fatalf("attr count: got %d want %d", len(got), len(want))
	}
	for name, wt := range want {
		gt, ok := got[name]
		if !ok {
			t.Fatalf("missing attribute %q on null list object type", name)
		}
		if !gt.Equal(wt) {
			t.Fatalf("attribute %q: got %#v want %#v", name, gt, wt)
		}
	}
}

func TestElasticsearchConnectionBlockObjectAttrTypes_returnsCopy(t *testing.T) {
	t.Parallel()

	first := elasticsearchConnectionBlockObjectAttrTypes()
	second := elasticsearchConnectionBlockObjectAttrTypes()

	require.Equal(t, first, second)
	first["mutated"] = types.BoolType
	require.NotEqual(t, first, second)
}

func TestElasticsearchConnectionFallbackAttrTypes_matchGetEsFWConnectionBlock(t *testing.T) {
	t.Parallel()

	block := GetEsFWConnectionBlock()
	lb, ok := block.(fwschema.ListNestedBlock)
	require.True(t, ok, "GetEsFWConnectionBlock must return a ListNestedBlock")

	want, err := fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
	require.NoError(t, err)

	got := elasticsearchConnectionBlockObjectAttrTypesFallback()
	require.Equal(t, want, got)
}

func TestElasticsearchConnectionBlocks_includeCAFingerprintAttribute(t *testing.T) {
	t.Parallel()

	require.Contains(t, fwConnectionBlockAttributeNames(GetEsFWConnectionBlock()), attrCAFingerprint)
	require.Contains(t, ephemeralConnectionBlockAttributeNames(GetEsEphemeralConnectionBlock()), attrCAFingerprint)
	require.Contains(t, actionConnectionBlockAttributeNames(GetEsActionConnectionBlock()), attrCAFingerprint)
}

func TestElasticsearchConnectionBlocks_tlsTrustAttributesHaveMatchingValidatorCounts(t *testing.T) {
	t.Parallel()

	managed := tlsTrustAttributeValidatorCounts{
		caFile:        countFWStringValidators(GetEsFWConnectionBlock(), attrCAFile),
		caData:        countFWStringValidators(GetEsFWConnectionBlock(), attrCAData),
		caFingerprint: countFWStringValidators(GetEsFWConnectionBlock(), attrCAFingerprint),
	}
	ephemeral := tlsTrustAttributeValidatorCounts{
		caFile:        countEphemeralStringValidators(GetEsEphemeralConnectionBlock(), attrCAFile),
		caData:        countEphemeralStringValidators(GetEsEphemeralConnectionBlock(), attrCAData),
		caFingerprint: countEphemeralStringValidators(GetEsEphemeralConnectionBlock(), attrCAFingerprint),
	}
	action := tlsTrustAttributeValidatorCounts{
		caFile:        countActionStringValidators(GetEsActionConnectionBlock(), attrCAFile),
		caData:        countActionStringValidators(GetEsActionConnectionBlock(), attrCAData),
		caFingerprint: countActionStringValidators(GetEsActionConnectionBlock(), attrCAFingerprint),
	}

	for _, counts := range []tlsTrustAttributeValidatorCounts{managed, ephemeral, action} {
		require.GreaterOrEqual(t, counts.caFile, 1)
		require.GreaterOrEqual(t, counts.caData, 1)
		require.GreaterOrEqual(t, counts.caFingerprint, 1)
	}

	require.Equal(t, managed, ephemeral)
	require.Equal(t, managed, action)
}

func TestKibanaConnectionNullList_objectMatchesGetKbFWConnectionBlock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	block := GetKbFWConnectionBlock()
	lb, ok := block.(fwschema.ListNestedBlock)
	if !ok {
		t.Fatal("GetKbFWConnectionBlock must return a ListNestedBlock")
	}
	want, err := fwNestedBlockAttributesToAttrTypes(lb.NestedObject.Attributes)
	if err != nil {
		t.Fatalf("fwNestedBlockAttributesToAttrTypes: %v", err)
	}

	list := KibanaConnectionNullList()
	listType, ok := list.Type(ctx).(types.ListType)
	if !ok {
		t.Fatalf("KibanaConnectionNullList.Type: got %T", list.Type(ctx))
	}
	objType, ok := listType.ElementType().(types.ObjectType)
	if !ok {
		t.Fatalf("list element type: got %T", listType.ElementType())
	}
	got := objType.AttrTypes

	if len(got) != len(want) {
		t.Fatalf("attr count: got %d want %d", len(got), len(want))
	}
	for name, wt := range want {
		gt, ok := got[name]
		if !ok {
			t.Fatalf("missing attribute %q on null list object type", name)
		}
		if !gt.Equal(wt) {
			t.Fatalf("attribute %q: got %#v want %#v", name, gt, wt)
		}
	}
}
