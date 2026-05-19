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

package githubx

import (
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/semver"
)

const headRefCompare = "HEAD"

// ListCommitSHAs executes `git log --format=%H <range>` like changelog-engine-factory.js gatherMergedPRRecordsForRange.
func ListCommitSHAs(exec semver.Execer, compareRange string) ([]string, error) {
	rng := strings.TrimSpace(compareRange)
	if rng == "" {
		rng = headRefCompare
	}
	out, err := exec.Run("git", "log", "--format=%H", rng)
	if err != nil {
		return nil, err
	}
	var lines []string
	for ln := range strings.SplitSeq(string(out), "\n") {
		s := strings.TrimSpace(ln)
		if s != "" {
			lines = append(lines, s)
		}
	}
	return lines, nil
}
