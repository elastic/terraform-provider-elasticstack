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

package agentbuilder

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SpaceIDSetter is implemented by Agent Builder plan models that store a
// mutable space_id field, letting [SetWriteSpaceID] write it back without
// each model needing its own copy of that assignment.
type SpaceIDSetter interface {
	SetSpaceID(types.String)
}

// SetWriteSpaceID is a populate callback for entitycore.SimpleKibanaCreate
// and entitycore.SimpleKibanaUpdate shared by the Agent Builder packages
// (agent, skill, tool, workflow): it records the write-time space ID on the
// plan so the envelope's read-after-write step operates on the resolved
// space, and otherwise ignores the API response. Resources with extra
// post-write steps (for example agentbuilderworkflow, which also captures a
// generated workflow ID) should call this from their own populate callback
// rather than replacing it outright.
func SetWriteSpaceID[M SpaceIDSetter, R any](model M, _ context.Context, spaceID string, _ *R) diag.Diagnostics {
	model.SetSpaceID(types.StringValue(spaceID))
	return nil
}
