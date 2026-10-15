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

package kibanaoapi

import "github.com/hashicorp/terraform-plugin-framework/diag"

// PageFetcher fetches one 1-based page of items from a Kibana list endpoint.
// total is the number of items available across all pages, as reported by the
// backend on that call.
type PageFetcher[T any] func(page int) (items []T, total int, diags diag.Diagnostics)

// CollectAllPages repeatedly calls fetch, starting at page 1, accumulating items
// until every item reported by the backend (via total on the first page) has
// been collected or a page comes back empty.
func CollectAllPages[T any](fetch PageFetcher[T]) ([]T, diag.Diagnostics) {
	var collected []T
	total := 0

	for page := 1; ; page++ {
		items, reportedTotal, diags := fetch(page)
		if diags.HasError() {
			return nil, diags
		}

		collected = append(collected, items...)
		if page == 1 {
			total = reportedTotal
		}

		if len(items) == 0 || len(collected) >= total {
			break
		}
	}

	return collected, nil
}

// FindFirstPage walks pages calling fetch, starting at page 1, and returns the
// first item for which match returns true. It stops as soon as a match is
// found, once every item reported by the backend has been seen, or when a page
// comes back empty.
func FindFirstPage[T any](fetch PageFetcher[T], match func(T) bool) (T, bool, diag.Diagnostics) {
	var zero T
	seen := 0
	total := 0

	for page := 1; ; page++ {
		items, reportedTotal, diags := fetch(page)
		if diags.HasError() {
			return zero, false, diags
		}
		if page == 1 {
			total = reportedTotal
		}

		for _, item := range items {
			if match(item) {
				return item, true, nil
			}
		}

		seen += len(items)
		if len(items) == 0 || seen >= total {
			break
		}
	}

	return zero, false, nil
}
