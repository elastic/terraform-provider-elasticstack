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

package role

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestFromAPIModel_PreservesEmptyStringDescriptionWhenAPIIsNull(t *testing.T) {
	ctx := context.Background()

	d := Data{
		Name:        types.StringValue("role-a"),
		Description: types.StringValue(""),
	}

	diags := d.fromAPIModel(ctx, &models.Role{
		Name:        "role-a",
		Description: nil,
	})
	require.False(t, diags.HasError(), "unexpected diags: %#v", diags)

	require.False(t, d.Description.IsNull())
	require.Empty(t, d.Description.ValueString())
}
