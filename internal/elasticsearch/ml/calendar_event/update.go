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

package calendar_event

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// updateCalendarEventNoOp is the envelope update callback: calendar events have no in-place
// update API (attribute changes use RequiresReplace), but the Elasticsearch envelope still
// invokes Update when only nested blocks such as elasticsearch_connection change. Returning
// the plan unchanged lets read refresh state without failing the apply.
func updateCalendarEventNoOp(_ context.Context, _ *clients.ElasticsearchScopedClient, _ string, plan CalendarEventTFModel, _ CalendarEventTFModel) (CalendarEventTFModel, diag.Diagnostics) {
	return plan, nil
}
