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

package ml

import (
	"strings"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// SplitCalendarResourcePath splits a <calendar_id>/<sub_resource_id> path
// segment and returns both parts. subResourceLabel is used only in the error
// message (e.g. "<event_id>" or "<job_id>").
func SplitCalendarResourcePath(resourcePath, subResourceLabel string) (calendarID, subResourceID string, diags fwdiags.Diagnostics) {
	parts := strings.SplitN(resourcePath, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		diags.AddError("Invalid ID format", "Expected resource segment format: <calendar_id>/"+subResourceLabel)
		return "", "", diags
	}
	return parts[0], parts[1], diags
}
