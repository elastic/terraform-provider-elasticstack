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

package entitycore

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDAttribute_DefaultDescription(t *testing.T) {
	t.Parallel()
	attr := IDAttribute()

	assert.Equal(t, defaultIDAttributeDescription, attr.MarkdownDescription)
	assert.True(t, attr.Computed)
	assert.False(t, attr.Required)
	assert.False(t, attr.Optional)
	require.Len(t, attr.PlanModifiers, 1)
	assert.IsType(t, stringplanmodifier.UseStateForUnknown(), attr.PlanModifiers[0])
}
