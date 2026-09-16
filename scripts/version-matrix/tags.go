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
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-github/v89/github"
)

var gaTagPattern = regexp.MustCompile(`^v(8|9)\.\d+\.\d+$`)

type minorKey struct {
	major int
	minor int
}

const (
	elasticsearchOwner = "elastic"
	elasticsearchRepo  = "elasticsearch"
)

func ListElasticsearchTags(ctx context.Context, client *github.Client) ([]string, error) {
	opts := &github.ListOptions{PerPage: 100}
	var names []string
	for {
		tags, resp, err := client.Repositories.ListTags(ctx, elasticsearchOwner, elasticsearchRepo, opts)
		if err != nil {
			return nil, fmt.Errorf("list elasticsearch tags: %w", err)
		}
		for _, tag := range tags {
			if tag == nil || tag.Name == nil {
				continue
			}
			names = append(names, *tag.Name)
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return names, nil
}

func LatestPatchPerMinor(tags []string) []string {
	best := make(map[minorKey]int)
	for _, tag := range tags {
		if !gaTagPattern.MatchString(tag) {
			continue
		}
		major, minor, patch, ok := parseGATag(tag)
		if !ok {
			continue
		}
		key := minorKey{major: major, minor: minor}
		if prev, exists := best[key]; !exists || patch > prev {
			best[key] = patch
		}
	}

	out := make([]string, 0, len(best))
	for key, patch := range best {
		out = append(out, fmt.Sprintf("%d.%d.%d", key.major, key.minor, patch))
	}
	sortVersionStrings(out)
	return out
}

// splitVersionComponents strips a leading "v" and splits the remainder on ".",
// the trim/split shape shared by parseGATag and parseLooseVersion.
func splitVersionComponents(v string) []string {
	trimmed := strings.TrimPrefix(v, "v")
	return strings.Split(trimmed, ".")
}

// atoiOrZero converts s to an int, defaulting to 0 for invalid input.
func atoiOrZero(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseGATag(tag string) (major, minor, patch int, ok bool) {
	parts := splitVersionComponents(tag)
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	ints := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return 0, 0, 0, false
		}
		ints[i] = n
	}
	return ints[0], ints[1], ints[2], true
}

func sortVersionStrings(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		return versionLess(versions[i], versions[j])
	})
}

func versionLess(a, b string) bool {
	am, an, ap := parseLooseVersion(a)
	bm, bn, bp := parseLooseVersion(b)
	if am != bm {
		return am < bm
	}
	if an != bn {
		return an < bn
	}
	return ap < bp
}

func parseLooseVersion(v string) (major, minor, patch int) {
	v = strings.TrimSuffix(v, "-SNAPSHOT")
	parts := splitVersionComponents(v)
	if len(parts) > 0 {
		major = atoiOrZero(parts[0])
	}
	if len(parts) > 1 {
		minor = atoiOrZero(parts[1])
	}
	if len(parts) > 2 {
		patch = atoiOrZero(parts[2])
	}
	return major, minor, patch
}
