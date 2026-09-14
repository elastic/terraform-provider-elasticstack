// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Elasticsearch B.V.
// licenses this file to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and limitations
// under the License.

package monitor

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"
)

type minVersionClient struct {
	minVersion *version.Version
	supported  bool
}

func (c *minVersionClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, diag.Diagnostics) {
	c.minVersion = minVersion
	return c.supported, nil
}

func TestEnforceKibanaSpacesVersionRejectsUnsupportedServer(t *testing.T) {
	client := &minVersionClient{}

	diags := enforceKibanaSpacesVersion(
		context.Background(),
		client,
		typeutils.StringsToListMust([]string{"security"}),
	)

	require.True(t, diags.HasError())
	require.Equal(t, MinKibanaSpacesVersion, client.minVersion)
	require.Equal(t, "Unsupported version for `kibana_spaces` attribute", diags.Errors()[0].Summary())
	require.Contains(t, diags.Errors()[0].Detail(), MinKibanaSpacesVersion.String())
}
