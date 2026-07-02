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

// Package dashboardacctest exposes shared helpers used by Kibana Dashboard
// acceptance tests across multiple packages (dashboard root, panel/*).
package dashboardacctest

import "github.com/hashicorp/go-version"

// MinDashboardAPISupport is the lowest Elastic Stack version known to expose the
// Kibana Dashboard API surface that these acceptance tests exercise.
var MinDashboardAPISupport = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))

// MinControlByFieldEsqlUnionSupport is the lowest Elastic Stack version whose
// Kibana server registers the `options_list_control` / `range_slider_control`
// panel config schemas as a `values_source`-discriminated union (Field vs
// ES|QL variant) — i.e. the wire shape the by_field/by_esql restructure (see
// the esql-control-variants OpenSpec change) targets.
//
// Confirmed by reading the bundled Kibana server source directly:
//   - 9.4.0: @kbn/controls-schemas's `optionsListDSLControlSchema` and
//     `rangeSliderControlSchema` are each a single fixed object schema with NO
//     `values_source` property at all (`unknowns` is not "allow", so a Field
//     write that included `values_source` would be rejected with "Additional
//     properties are not allowed ('values_source' was unexpected)"). There is
//     no ES|QL variant of either schema in 9.4.0.
//   - 9.5.0-SNAPSHOT: `optionsListDSLControlSchema` became
//     `schema.discriminatedUnion('values_source', [esql, field])`, matching
//     the `KibanaHTTPAPIsKbnControlsSchemasOptionsListDslControlSchema{Field,Esql}`
//     shapes this provider's optionslist/rangeslider packages already convert
//     to/from.
//
// by_field writes deliberately omit `values_source` on the wire (see
// buildFieldConfig in each package) precisely so they remain compatible with
// every Kibana version this resource supports, including < 9.4.0. Only the
// by_esql branch is gated on this constant: it is a genuinely new API surface
// that does not exist on Kibana servers below 9.5.0-SNAPSHOT.
var MinControlByFieldEsqlUnionSupport = version.Must(version.NewVersion("9.5.0-SNAPSHOT"))
