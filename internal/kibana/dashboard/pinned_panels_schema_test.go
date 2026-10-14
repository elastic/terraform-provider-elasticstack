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

package dashboard

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_pinnedPanelsNestedObject_widthAndGrowAttributes(t *testing.T) {
	t.Parallel()
	attrs := pinnedPanelsNestedObject().Attributes

	width, ok := attrs["width"].(schema.StringAttribute)
	require.True(t, ok, "width must be an optional string attribute")
	require.True(t, width.Optional)

	grow, ok := attrs["grow"].(schema.BoolAttribute)
	require.True(t, ok, "grow must be an optional bool attribute")
	require.True(t, grow.Optional)

	validate := func(v string) bool {
		resp := &validator.StringResponse{}
		for _, val := range width.Validators {
			val.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(v)}, resp)
		}
		return !resp.Diagnostics.HasError()
	}
	for _, v := range []string{"small", "medium", "large"} {
		require.True(t, validate(v), v)
	}
	require.False(t, validate("huge"))
}
