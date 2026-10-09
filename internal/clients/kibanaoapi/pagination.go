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

// FetchPageFunc fetches a single 1-based page of results from a Kibana "find"-style
// paginated endpoint. It returns the items found on that page and the total number of
// items the API reports as matching the query (as reported on the first page).
type FetchPageFunc[T any] func(page float32) (items []T, total float32, diags diag.Diagnostics)

// CollectAllPages pages through fetchPage starting at page 1, accumulating every item,
// until the accumulated item count reaches the total reported on the first page or a
// page comes back with no items.
func CollectAllPages[T any](fetchPage FetchPageFunc[T]) ([]T, diag.Diagnostics) {
	var (
		collected []T
		page      float32 = 1
		total     float32
	)

	for {
		items, pageTotal, diags := fetchPage(page)
		if diags.HasError() {
			return nil, diags
		}

		collected = append(collected, items...)
		if page == 1 {
			total = pageTotal
		}

		if len(collected) >= int(total) || len(items) == 0 {
			break
		}
		page++
	}

	return collected, nil
}
