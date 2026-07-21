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

package managedintegration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConflictHintDiagnostics covers a bug found in review: this function
// used to take the full diag.Diagnostics from fleetclient.DeleteAgentlessPolicy
// and infer a conflict by pattern-matching diagutil.ReportUnknownHTTPError's
// generated summary text ("... got HTTP 409 ..."), which was brittle against
// wording changes or a switch to a different error-reporting helper (e.g.
// diagutil.ReportKibanaBoomHTTPError, whose summary is caller-supplied and
// might not contain "HTTP 409" at all). fleetclient.DeleteAgentlessPolicy now
// derives the conflict signal from the final HTTP status code observed across
// retries and reports it as a plain bool (see internal/clients/fleet/
// agentless_policy_compat.go and TestDeleteAgentlessPolicy/
// max_retries_exhausted_returns_error and
// transport_error_after_409_resets_is_conflict), so this function no longer needs to
// inspect diagnostic text at all -- it is now a pure "build the hint" helper
// that deleteAgentlessPolicy only calls once it already knows, authoritatively,
// that the delete failed with a conflict.
func TestConflictHintDiagnostics(t *testing.T) {
	t.Parallel()

	hint := conflictHintDiagnostics()
	require.Len(t, hint, 1)
	assert.True(t, hint.HasError())
	assert.Contains(t, hint[0].Summary(), "conflict")
	assert.Contains(t, hint[0].Detail(), "force_delete")
	assert.Contains(t, hint[0].Detail(), "force=true")
}
