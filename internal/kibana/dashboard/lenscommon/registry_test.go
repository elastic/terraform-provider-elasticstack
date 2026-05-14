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

package lenscommon

// ForType opaque-alias assertions against the live converter registry are in
// registry_aliases_external_test.go as package lenscommon_test to avoid an import cycle
// (panel/lens* packages import lenscommon).

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func saveConvertersRegistry(t *testing.T) {
	t.Helper()
	prev := convertersByType
	t.Cleanup(func() { convertersByType = prev })
	convertersByType = nil
}

func saveSliceAligners(t *testing.T) {
	t.Helper()
	prev := sliceAligners
	t.Cleanup(func() { sliceAligners = prev })
	sliceAligners = nil
}

type fakeConverter struct {
	vizType       string
	handlesBlocks func(blocks *models.LensByValueChartBlocks) bool
}

func (f fakeConverter) VizType() string { return f.vizType }

func (f fakeConverter) HandlesBlocks(blocks *models.LensByValueChartBlocks) bool {
	if f.handlesBlocks != nil {
		return f.handlesBlocks(blocks)
	}
	return false
}

func (fakeConverter) SchemaAttribute() schema.Attribute {
	return schema.StringAttribute{Optional: true}
}

func (fakeConverter) PopulateFromAttributes(context.Context, Resolver, *models.LensByValueChartBlocks, kbapi.KbnDashboardPanelTypeVisConfig0) diag.Diagnostics {
	return nil
}

func (fakeConverter) BuildAttributes(*models.LensByValueChartBlocks, Resolver) (kbapi.KbnDashboardPanelTypeVisConfig0, diag.Diagnostics) {
	return kbapi.KbnDashboardPanelTypeVisConfig0{}, nil
}

func (fakeConverter) AlignStateFromPlan(context.Context, *models.LensByValueChartBlocks, *models.LensByValueChartBlocks) {
}

func (fakeConverter) PopulateJSONDefaults(attrs map[string]any) map[string]any {
	return attrs
}

func TestForType_unknownReturnsNil(t *testing.T) {
	saveConvertersRegistry(t)
	require.Nil(t, ForType("missing"))
}

func TestRegister_ForType_roundTrip(t *testing.T) {
	saveConvertersRegistry(t)
	c := fakeConverter{vizType: "gauge"}
	Register(c)
	got := ForType("gauge")
	require.NotNil(t, got)
	require.Equal(t, "gauge", got.VizType())
}

func TestRegister_sameVizTypeOverwritesSilently(t *testing.T) {
	saveConvertersRegistry(t)
	var marker string
	first := fakeConverter{
		vizType: "pie",
		handlesBlocks: func(*models.LensByValueChartBlocks) bool {
			marker = "first"
			return true
		},
	}
	second := fakeConverter{
		vizType: "pie",
		handlesBlocks: func(*models.LensByValueChartBlocks) bool {
			marker = "second"
			return true
		},
	}
	Register(first)
	Register(second)
	got := ForType("pie")
	require.NotNil(t, got)
	_ = got.HandlesBlocks(&models.LensByValueChartBlocks{})
	require.Equal(t, "second", marker, "later Register should replace the converter for the same VizType() without error")
}

func TestFirstForBlocks_nilBlocks(t *testing.T) {
	saveConvertersRegistry(t)
	Register(fakeConverter{vizType: "xy"})
	c, ok := FirstForBlocks(nil)
	require.False(t, ok)
	require.Nil(t, c)
}

func TestFirstForBlocks_returnsFirstMatchingConverterInSortedOrder(t *testing.T) {
	saveConvertersRegistry(t)
	blocks := &models.LensByValueChartBlocks{}
	matchAny := func(*models.LensByValueChartBlocks) bool { return true }
	Register(fakeConverter{vizType: "zebra", handlesBlocks: matchAny})
	Register(fakeConverter{vizType: "alpha", handlesBlocks: matchAny})

	got, ok := FirstForBlocks(blocks)
	require.True(t, ok)
	require.Equal(t, "alpha", got.VizType())
}

func TestAll_sortedByVizType(t *testing.T) {
	saveConvertersRegistry(t)
	Register(fakeConverter{vizType: "m"})
	Register(fakeConverter{vizType: "a"})
	Register(fakeConverter{vizType: "z"})
	all := All()
	require.Len(t, all, 3)
	require.Equal(t, "a", all[0].VizType())
	require.Equal(t, "m", all[1].VizType())
	require.Equal(t, "z", all[2].VizType())
}
