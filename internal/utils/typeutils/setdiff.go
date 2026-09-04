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

package typeutils

import "sort"

// DiffStringSlices partitions after relative to before into added and
// removed elements, keyed by string equality. Both returned slices are
// sorted for deterministic output.
func DiffStringSlices(before, after []string) (added, removed []string) {
	beforeSet := make(map[string]struct{}, len(before))
	for _, v := range before {
		beforeSet[v] = struct{}{}
	}
	afterSet := make(map[string]struct{}, len(after))
	for _, v := range after {
		afterSet[v] = struct{}{}
	}

	for v := range afterSet {
		if _, ok := beforeSet[v]; !ok {
			added = append(added, v)
		}
	}
	for v := range beforeSet {
		if _, ok := afterSet[v]; !ok {
			removed = append(removed, v)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}
