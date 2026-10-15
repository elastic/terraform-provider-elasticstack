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

package rolloverindexseries

import _ "embed"

//go:embed resource-description.md
var resourceDescription string

const (
	descID = "Internal identifier in the form `<cluster_uuid>/<rollover_alias>`. " +
		"It stays the same when the effective writer or alias membership changes."
	descPrefix = "Literal prefix used only to name the bootstrap index `<prefix>-000001`. " +
		"Changing it forces replacement. Membership and deletion follow `rollover_alias`, not this prefix."
	descRolloverAlias = "Dedicated ordinary-index alias. It is the series identity and the only membership boundary. " +
		"Changing it forces replacement."
	descDeletionProtection = "Whether Terraform may destroy this series. Defaults to `true`. " +
		"While enabled, destroy is rejected and no index is deleted. " +
		"Set it to `false` and apply that change before a destroy. " +
		"Destroy then deletes every index currently attached to `rollover_alias`, including indices outside `prefix`."
	descWriteIndex = "Concrete name of the effective write index. Null when the alias has no effective writer."
	descIndices    = "Concrete indices currently attached to `rollover_alias`, including members that do not match `prefix`."
)
