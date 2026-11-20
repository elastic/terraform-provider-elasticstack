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

package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// AlreadyExistsDiagnostic builds the error returned when Create finds that a
// noun (for example "rule") with importID already exists in the target space:
// Terraform does not adopt unmanaged objects, so the user has to import it.
func AlreadyExistsDiagnostic(noun, importID string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Alerting v2 "+noun+" already exists",
		alreadyExistsDetail(noun, importID),
	)
}

// AlreadyReplacedDiagnostic is the AlreadyExistsDiagnostic variant for the
// race in which the object appeared between Create's existence check and its
// PUT, so the PUT has already replaced the unmanaged object server-side.
func AlreadyReplacedDiagnostic(noun, importID string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Alerting v2 "+noun+" already exists",
		alreadyExistsDetail(noun, importID)+
			" The "+noun+" appeared after the existence check, so this request has already replaced its configuration server-side.",
	)
}

func alreadyExistsDetail(noun, importID string) string {
	return "A " + noun + " with id \"" + importID + "\" already exists and is not managed by this resource. " +
		"Run `terraform import` with id \"" + importID + "\" to manage it, or choose a different id."
}
