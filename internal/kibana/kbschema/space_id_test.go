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

package kbschema

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/stretchr/testify/require"
)

func TestResourceSpaceIDAttributeNoDefault(t *testing.T) {
	attr := ResourceSpaceIDAttributeNoDefault(
		stringplanmodifier.UseStateForUnknown(),
		stringplanmodifier.RequiresReplace(),
	)

	require.True(t, attr.Optional)
	require.True(t, attr.Computed)
	require.Nil(t, attr.Default)
	require.Equal(t, spaceIDDescription, attr.MarkdownDescription)
	require.Len(t, attr.PlanModifiers, 2)
}

func TestResourceSpaceIDAttributeNoDefault_NoModifiers(t *testing.T) {
	attr := ResourceSpaceIDAttributeNoDefault()

	require.Nil(t, attr.Default)
	require.Empty(t, attr.PlanModifiers)
}

func TestResourceSpaceIDAttributeHasDefaultUnlikeNoDefaultVariant(t *testing.T) {
	withDefault := ResourceSpaceIDAttribute()
	noDefault := ResourceSpaceIDAttributeNoDefault(stringplanmodifier.RequiresReplace())

	require.NotNil(t, withDefault.Default)
	require.Nil(t, noDefault.Default)
}
