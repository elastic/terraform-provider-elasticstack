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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestPreserveSloInstanceID_noPrior_adoptsConcreteAPIValue(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("instance-1"), false, types.StringNull())
	assert.Equal(t, types.StringValue("instance-1"), got)
}

func TestPreserveSloInstanceID_noPrior_normalizesWildcardToNull(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("*"), false, types.StringNull())
	assert.True(t, got.IsNull())
}

func TestPreserveSloInstanceID_noPrior_nilAPIIsNull(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(nil, false, types.StringNull())
	assert.True(t, got.IsNull())
}

func TestPreserveSloInstanceID_unknownPrior_staysNullEvenForConcreteAPIValue(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("instance-1"), true, types.StringNull())
	assert.True(t, got.IsNull(), "REQ-009: an unconfigured field must not adopt drift from the API")
}

func TestPreserveSloInstanceID_unknownPrior_staysNullForWildcard(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("*"), true, types.StringUnknown())
	assert.True(t, got.IsNull())
}

func TestPreserveSloInstanceID_knownPrior_adoptsConcreteAPIValue(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("instance-1"), true, types.StringValue("instance-1"))
	assert.Equal(t, types.StringValue("instance-1"), got)
}

func TestPreserveSloInstanceID_knownPrior_roundTripsExplicitWildcard(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(new("*"), true, types.StringValue("*"))
	assert.Equal(t, types.StringValue("*"), got, "a practitioner who explicitly set \"*\" must see it round-trip, not be nulled")
}

func TestPreserveSloInstanceID_knownPrior_apiOmitsValue_isNull(t *testing.T) {
	t.Parallel()
	got := PreserveSloInstanceID(nil, true, types.StringValue("instance-1"))
	assert.True(t, got.IsNull(), "a known prior is refreshed from the API like any other optional scalar field")
}
