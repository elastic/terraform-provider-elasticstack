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

package monitor

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestMonitorVersionRequirementsLabelsKnown(t *testing.T) {
	t.Parallel()

	model := tfModelV0{
		Labels: types.MapValueMust(types.StringType, map[string]attr.Value{"environment": types.StringValue("production")}),
	}

	reqs, diags := model.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Len(t, reqs, 1)
	require.NotNil(t, reqs[0].AttributePath)
	require.Equal(t, "labels", reqs[0].AttributePath.String())
	require.Equal(t, MinLabelsVersion.String(), reqs[0].MinVersion.String())
}

func TestMonitorVersionRequirementsLabelsNotConfigured(t *testing.T) {
	t.Parallel()

	reqs, diags := tfModelV0{}.GetVersionRequirements(context.Background())
	require.False(t, diags.HasError())
	require.Empty(t, reqs)
}
