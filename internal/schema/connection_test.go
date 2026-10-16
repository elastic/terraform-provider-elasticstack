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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestElasticsearchConnectionBlockObjectAttrTypes_returnsCopy(t *testing.T) {
	t.Parallel()

	first := elasticsearchConnectionBlockObjectAttrTypes()
	second := elasticsearchConnectionBlockObjectAttrTypes()

	require.Equal(t, first, second)
	first["mutated"] = types.BoolType
	require.NotEqual(t, first, second)
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

func TestElasticsearchConnectionBlocks_allAttributesPresent(t *testing.T) {
	t.Parallel()

	wantAttrs := []string{
		attrUsername, attrPassword, attrAPIKey, attrBearerToken,
		attrESClientAuthentication, attrEndpoints, attrHeaders,
		attrInsecure, attrCAFile, attrCAData, attrCAFingerprint,
		attrCertFile, attrKeyFile, attrCertData, attrKeyData,
	}

	managed := fwConnectionBlockAttributeNames(GetEsFWConnectionBlock())
	ephemeral := ephemeralConnectionBlockAttributeNames(GetEsEphemeralConnectionBlock())
	action := actionConnectionBlockAttributeNames(GetEsActionConnectionBlock())

	for _, attr := range wantAttrs {
		if _, ok := managed[attr]; !ok {
			t.Errorf("managed connection block missing attribute %q", attr)
		}
		if _, ok := ephemeral[attr]; !ok {
			t.Errorf("ephemeral connection block missing attribute %q", attr)
		}
		if _, ok := action[attr]; !ok {
			t.Errorf("action connection block missing attribute %q", attr)
		}
	}
}
