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

package elasticsearch

import (
	"fmt"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// LookupOrNotFoundDiag returns the value keyed by key in m, or a diagnostic
// reporting that key was not found among resourceLabel in the cluster. This
// covers Get* client functions whose response is a map keyed by the
// resource's identifier.
func LookupOrNotFoundDiag[T any](m map[string]T, key, resourceLabel string) (*T, fwdiags.Diagnostics) {
	if v, ok := m[key]; ok {
		return &v, nil
	}
	return nil, notFoundDiag(key, resourceLabel)
}

// SingleOrNotFoundDiag returns the sole element of items, or a diagnostic
// reporting that key was not found among resourceLabel in the cluster if
// items does not contain exactly one element. This covers Get* client
// functions whose response is a slice expected to contain exactly one
// matching resource.
func SingleOrNotFoundDiag[T any](items []T, key, resourceLabel string) (*T, fwdiags.Diagnostics) {
	if len(items) != 1 {
		return nil, notFoundDiag(key, resourceLabel)
	}
	item := items[0]
	return &item, nil
}

// notFoundDiag builds the shared "get single named resource" not-found
// diagnostic, keeping the summary/detail wording consistent across the
// Get* client functions that use LookupOrNotFoundDiag/SingleOrNotFoundDiag.
func notFoundDiag(key, resourceLabel string) fwdiags.Diagnostics {
	return fwdiags.Diagnostics{
		fwdiags.NewErrorDiagnostic(
			fmt.Sprintf("Unable to find %s in the cluster", resourceLabel),
			fmt.Sprintf(`Unable to find "%s" %s in the cluster`, key, resourceLabel),
		),
	}
}
