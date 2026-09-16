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

import "github.com/hashicorp/go-version"

// MinKibanaAPISupport95 is the lowest Kibana version whose Dashboard API accepts the panel types
// that shipped in the 9.5 line (aiops_change_point_chart, aiops_log_rate_analysis,
// aiops_pattern_analysis, apm_service_map, field_stats_table, links, ml_anomaly_charts,
// ml_anomaly_swimlane, ml_single_metric_viewer). Empirical testing showed Kibana 9.4.0 rejects
// these panel types (HTTP 400); 9.5.0-SNAPSHOT accepts them.
var MinKibanaAPISupport95 = version.Must(version.NewVersion("9.5.0-SNAPSHOT"))
