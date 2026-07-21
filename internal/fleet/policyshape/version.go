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

package policyshape

import "github.com/hashicorp/go-version"

// MinVersionCondition is the minimum Kibana version that accepts the
// `condition` field on package-policy inputs/streams. Verified empirically
// against a 9.5.0-SNAPSHOT Kibana: 9.4.0 and 9.4.3 both reject it with an
// "Additional properties are not allowed" 400.
//
// This constant lives in policyshape (rather than in each resource package
// that surfaces `condition`) because `condition` itself is part of the
// InputType/StreamType shape this package owns (see AttrCondition in
// attribute_types.go and InputModel/InputStreamModel in models.go): every
// caller of that shared shape needs the same minimum version, and duplicating
// the literal per resource package would risk the two silently drifting out
// of sync on a future Kibana change. Originally introduced (as an
// integration_policy-local var) for internal/fleet/integration_policy; see
// that package's design.md Open Question 4 resolution for the original
// empirical investigation. internal/fleet/managedintegration reuses this same
// constant for the identical gating requirement on its own `condition`
// attributes.
var MinVersionCondition = version.Must(version.NewVersion("9.5.0"))
