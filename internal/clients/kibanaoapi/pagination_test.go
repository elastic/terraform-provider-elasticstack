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
	"github.com/stretchr/testify/require"
)

func TestCollectAllPages_collectsUntilTotalReached(t *testing.T) {
	t.Parallel()

	var pagesRequested []float32
	fetch := func(page float32) ([]string, float32, diag.Diagnostics) {
		pagesRequested = append(pagesRequested, page)
		switch page {
		case 1:
			return []string{"a", "b"}, 5, nil
		case 2:
			return []string{"c", "d"}, 5, nil
		case 3:
			return []string{"e"}, 5, nil
		default:
			t.Fatalf("unexpected page %v", page)
			return nil, 0, nil
		}
	}

	items, diags := kibanaoapi.CollectAllPages(fetch)
	require.False(t, diags.HasError())
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, items)
	require.Equal(t, []float32{1, 2, 3}, pagesRequested)
}

func TestCollectAllPages_stopsOnEmptyPage(t *testing.T) {
	t.Parallel()

	var pagesRequested int
	fetch := func(page float32) ([]string, float32, diag.Diagnostics) {
		pagesRequested++
		switch page {
		case 1:
			return []string{"a"}, 3, nil
		case 2:
			return nil, 3, nil
		default:
			t.Fatalf("unexpected page %v", page)
			return nil, 0, nil
		}
	}

	items, diags := kibanaoapi.CollectAllPages(fetch)
	require.False(t, diags.HasError())
	require.Equal(t, []string{"a"}, items)
	require.Equal(t, 2, pagesRequested)
}

func TestCollectAllPages_stopsOnFirstPageWhenTotalIsZero(t *testing.T) {
	t.Parallel()

	var pagesRequested int
	fetch := func(page float32) ([]string, float32, diag.Diagnostics) {
		pagesRequested++
		return nil, 0, nil
	}

	items, diags := kibanaoapi.CollectAllPages(fetch)
	require.False(t, diags.HasError())
	require.Empty(t, items)
	require.Equal(t, 1, pagesRequested)
}

func TestCollectAllPages_laterPageFailureDiscardsEarlierResults(t *testing.T) {
	t.Parallel()

	fetch := func(page float32) ([]string, float32, diag.Diagnostics) {
		if page == 1 {
			return []string{"a"}, 3, nil
		}
		return nil, 0, diag.Diagnostics{diag.NewErrorDiagnostic("boom", "page exploded")}
	}

	items, diags := kibanaoapi.CollectAllPages(fetch)
	require.True(t, diags.HasError())
	require.Nil(t, items)
}
