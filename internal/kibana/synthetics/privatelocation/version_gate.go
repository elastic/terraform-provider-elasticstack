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

package privatelocation

import (
	"strings"

	"github.com/hashicorp/go-version"
)

// MinVersionSpaceID is the minimum Elastic Stack version for using a non-default
// Kibana space with Synthetics private locations (space-scoped API paths for
// create, read, delete, and composite import).
var MinVersionSpaceID = version.Must(version.NewVersion("9.4.0-SNAPSHOT"))

// requiresSpaceIDMinVersion reports whether the stack must satisfy [MinVersionSpaceID]
// for the given effective Kibana space string used for API calls. The default space
// is represented by an empty string; the literal "default" is treated as default.
func requiresSpaceIDMinVersion(effectiveSpace string) bool {
	s := strings.TrimSpace(effectiveSpace)
	if s == "" || s == "default" {
		return false
	}
	return true
}
