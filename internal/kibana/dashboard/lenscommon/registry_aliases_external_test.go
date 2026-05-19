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

package lenscommon_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"

	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensdatatable"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensgauge"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensheatmap"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenslegacymetric"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensmetric"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensmosaic"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenspie"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensregionmap"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenstagcloud"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenstreemap"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenswaffle"
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensxy"
	"github.com/stretchr/testify/require"
)

func TestForType_opaqueAliasesResolveToCanonicalRegistryKey(t *testing.T) {
	tc := lenscommon.ForType("tagcloud")
	require.NotNil(t, tc)
	require.Equal(t, string(kbapi.TagcloudNoESQLTypeTagCloud), tc.VizType())

	dc := lenscommon.ForType("datatable")
	require.NotNil(t, dc)
	require.Equal(t, string(kbapi.DatatableNoESQLTypeDataTable), dc.VizType())

	require.NotNil(t, lenscommon.ForType(string(kbapi.TagcloudNoESQLTypeTagCloud)))
	require.NotNil(t, lenscommon.ForType(string(kbapi.DatatableNoESQLTypeDataTable)))

	require.Nil(t, lenscommon.ForType("not_a_registered_lens_chart_type_xyz"))
}
