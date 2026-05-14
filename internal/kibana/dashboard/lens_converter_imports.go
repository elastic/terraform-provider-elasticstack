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
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensdatatable"    // side-effect: register datatable VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensgauge"        // side-effect: register gauge VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensheatmap"      // side-effect: register heatmap VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenslegacymetric" // side-effect: register legacy_metric VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensmetric"       // side-effect: register metric VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensmosaic"       // side-effect: register mosaic VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenspie"          // side-effect: register pie VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensregionmap"    // side-effect: register region_map VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenstagcloud"     // side-effect: register tagcloud VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenstreemap"      // side-effect: register treemap VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lenswaffle"       // side-effect: register waffle VizConverter with lenscommon (dashboard-lens-contract).
	_ "github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/lensxy"           // side-effect: register xy VizConverter with lenscommon (dashboard-lens-contract).
)
