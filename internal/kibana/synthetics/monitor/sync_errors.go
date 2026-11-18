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

package monitor

import (
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const syncErrorsFollowUp = "Kibana retries syncing periodically. If the monitor does not run in these locations, check the Synthetics app or the Kibana logs."

// createSyncErrorsWarning describes push errors reported by the add-monitor
// API. Kibana returns the outcome of its most recent sync rather than the push
// for the monitor being created, so the errors may concern other monitors.
func createSyncErrorsWarning(name, id string, syncErrors []kibanaoapi.SyncError) diag.Diagnostics {
	if len(syncErrors) == 0 {
		return nil
	}
	return diag.Diagnostics{diag.NewWarningDiagnostic(
		"Kibana reported Synthetics Service sync errors",
		fmt.Sprintf(
			"The synthetics monitor %q (%s) was created and is tracked in Terraform state. "+
				"Kibana's response reported errors from syncing monitors to these Elastic-managed locations:\n\n%s\n\n"+
				"Kibana reports the outcome of its most recent sync, so these errors may relate to other monitors rather than this one. %s",
			name, id, formatSyncErrors(syncErrors), syncErrorsFollowUp,
		),
	)}
}

// updateSyncErrorsWarning describes push errors reported by the update-monitor
// API, which reports the push for the monitor being updated.
func updateSyncErrorsWarning(name, id string, syncErrors []kibanaoapi.SyncError) diag.Diagnostics {
	if len(syncErrors) == 0 {
		return nil
	}
	return diag.Diagnostics{diag.NewWarningDiagnostic(
		"Synthetics monitor saved, but syncing to Elastic-managed locations failed",
		fmt.Sprintf(
			"The synthetics monitor %q (%s) was saved in Kibana, but pushing it to the Synthetics Service failed for these Elastic-managed locations:\n\n%s\n\n%s",
			name, id, formatSyncErrors(syncErrors), syncErrorsFollowUp,
		),
	)}
}

func formatSyncErrors(syncErrors []kibanaoapi.SyncError) string {
	lines := make([]string, 0, len(syncErrors))
	for _, e := range syncErrors {
		location := e.LocationID
		if location == "" {
			location = "unknown location"
		}
		switch {
		case e.Reason != "" && e.Status != 0:
			lines = append(lines, fmt.Sprintf("- %s: %s (HTTP %d)", location, e.Reason, e.Status))
		case e.Reason != "":
			lines = append(lines, fmt.Sprintf("- %s: %s", location, e.Reason))
		case e.Status != 0:
			lines = append(lines, fmt.Sprintf("- %s: HTTP %d", location, e.Status))
		default:
			lines = append(lines, fmt.Sprintf("- %s (no details reported)", location))
		}
	}
	return strings.Join(lines, "\n")
}
