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

package main

import (
	"fmt"
	"regexp"
	"strings"
)

var versionPattern = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// introducedVersion extracts the Kibana version from an x-state value such as
// "Generally available in 8.15.0"; it is empty when the specification gives none.
func introducedVersion(xState string) string {
	return versionPattern.FindString(xState)
}

func renderTitle(g Group) string {
	return g.Resource + ": " + g.Field
}

func marker(g Group) string {
	return fmt.Sprintf("<!-- kibana-api-gap key=%s fp=%s -->", g.Key, g.Fingerprint)
}

func renderBody(g Group) string {
	var b strings.Builder
	fmt.Fprintf(&b, "The Terraform resource `%s` never uses the following Kibana API request fields under `%s`.\n\n", g.Resource, g.Field)
	for _, m := range g.Members {
		fmt.Fprintf(&b, "- `%s` (`%s`)", m.Field, m.Pointer)
		if version := introducedVersion(m.XState); version != "" {
			fmt.Fprintf(&b, ", introduced in %s", version)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nThe locations are JSON pointers into `generated/kbapi/oas-filtered.yaml`.\n\n")
	b.WriteString(marker(g))
	b.WriteString("\n")
	return b.String()
}
