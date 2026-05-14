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

package panelkit

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
)

// PanelConfigDescription builds MarkdownDescription for a mutually exclusive optional panel sibling
// (typed `*_config`, `config_json`, or other panel-level config blocks in the sibling list passed
// from dashboard registry init), documenting which other sibling attributes conflict with self.
func PanelConfigDescription(base, self string, names []string) string {
	others := make([]string, 0, len(names)-1)
	for _, name := range names {
		if name == self {
			continue
		}
		others = append(others, "`"+name+"`")
	}
	if len(others) == 0 {
		return base
	}
	return base + " Mutually exclusive with " + strings.Join(others, ", ") + "."
}

// SiblingTypedPanelConfigConflictPathsExcept returns path.Expression entries for sibling panel config
// blocks that must conflict-with the block named exceptName.
func SiblingTypedPanelConfigConflictPathsExcept(exceptName string, names []string) []path.Expression {
	paths := make([]path.Expression, 0, len(names)-1)
	for _, n := range names {
		if n == exceptName {
			continue
		}
		paths = append(paths, path.MatchRelative().AtParent().AtName(n))
	}
	return paths
}
