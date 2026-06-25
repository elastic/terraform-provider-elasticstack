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

package osquerysavedquery

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const prebuiltSavedQueryDiagnosticSummary = "Prebuilt Osquery saved query"

const prebuiltSavedQueryDiagnosticDetail = "Prebuilt Osquery saved queries are managed by the osquery_manager integration package " +
	"and cannot be managed by this resource. Use the elasticstack_kibana_osquery_saved_query data source to read this query."

func prebuiltGuardDiagnostic(prebuilt *bool) diag.Diagnostics {
	if prebuilt != nil && *prebuilt {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(prebuiltSavedQueryDiagnosticSummary, prebuiltSavedQueryDiagnosticDetail),
		}
	}
	return nil
}
