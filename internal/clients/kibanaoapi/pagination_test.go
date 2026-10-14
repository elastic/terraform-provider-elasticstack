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

package kibanaoapi_test

import (
	"testing"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectAllPages_collectsEveryPage(t *testing.T) {
	t.Parallel()

	pages := map[int][]int{
		1: {1, 2},
		2: {3, 4},
		3: {5},
	}
	var called []int

	items, diags := kibanaoapi.CollectAllPages(func(page int) ([]int, int, diag.Diagnostics) {
		called = append(called, page)
		return pages[page], 5, nil
	})

	require.False(t, diags.HasError())
	assert.Equal(t, []int{1, 2, 3, 4, 5}, items)
	assert.Equal(t, []int{1, 2, 3}, called)
}

func TestCollectAllPages_stopsOnEmptyPage(t *testing.T) {
	t.Parallel()

	var called int
	items, diags := kibanaoapi.CollectAllPages(func(page int) ([]int, int, diag.Diagnostics) {
		called++
		if page == 1 {
			return []int{1}, 10, nil
		}
		return nil, 10, nil
	})

	require.False(t, diags.HasError())
	assert.Equal(t, []int{1}, items)
	assert.Equal(t, 2, called)
}

func TestCollectAllPages_propagatesErrorAndDiscardsResults(t *testing.T) {
	t.Parallel()

	items, diags := kibanaoapi.CollectAllPages(func(page int) ([]int, int, diag.Diagnostics) {
		if page == 1 {
			return []int{1}, 10, nil
		}
		return nil, 0, diag.Diagnostics{diag.NewErrorDiagnostic("boom", "page failed")}
	})

	require.True(t, diags.HasError())
	assert.Nil(t, items)
}

func TestFindFirstPage_stopsAsSoonAsMatchFound(t *testing.T) {
	t.Parallel()

	pages := map[int][]int{
		1: {1, 2},
		2: {3, 4},
		3: {5, 6},
	}
	var called []int

	match, found, diags := kibanaoapi.FindFirstPage(func(page int) ([]int, int, diag.Diagnostics) {
		called = append(called, page)
		return pages[page], 6, nil
	}, func(item int) bool {
		return item == 3
	})

	require.False(t, diags.HasError())
	require.True(t, found)
	assert.Equal(t, 3, match)
	assert.Equal(t, []int{1, 2}, called)
}

func TestFindFirstPage_returnsNotFoundWhenExhausted(t *testing.T) {
	t.Parallel()

	match, found, diags := kibanaoapi.FindFirstPage(func(page int) ([]int, int, diag.Diagnostics) {
		if page == 1 {
			return []int{1, 2}, 2, nil
		}
		return nil, 2, nil
	}, func(item int) bool {
		return item == 99
	})

	require.False(t, diags.HasError())
	assert.False(t, found)
	assert.Zero(t, match)
}

func TestFindFirstPage_propagatesError(t *testing.T) {
	t.Parallel()

	_, found, diags := kibanaoapi.FindFirstPage(func(page int) ([]int, int, diag.Diagnostics) {
		return nil, 0, diag.Diagnostics{diag.NewErrorDiagnostic("boom", "page failed")}
	}, func(item int) bool {
		return true
	})

	require.True(t, diags.HasError())
	assert.False(t, found)
}
