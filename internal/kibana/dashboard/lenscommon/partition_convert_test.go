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

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestPartitionLegendFromTreemapLegend_Nil(t *testing.T) {
	t.Parallel()

	m := &models.PartitionLegendModel{}
	PartitionLegendFromTreemapLegend(m, nil)

	assert.Equal(t, types.BoolNull(), m.Nested)
	assert.Equal(t, types.StringNull(), m.Size)
	assert.Equal(t, types.Int64Null(), m.TruncateAfterLine)
	assert.Equal(t, types.StringNull(), m.Visible)
}

func TestPartitionLegendFromMosaicLegend_Nil(t *testing.T) {
	t.Parallel()

	m := &models.PartitionLegendModel{}
	PartitionLegendFromMosaicLegend(m, nil)

	assert.Equal(t, types.BoolNull(), m.Nested)
	assert.Equal(t, types.StringNull(), m.Size)
	assert.Equal(t, types.Int64Null(), m.TruncateAfterLine)
	assert.Equal(t, types.StringNull(), m.Visible)
}

func TestPartitionLegendFromTreemapAndMosaicLegend_FullyPopulated(t *testing.T) {
	t.Parallel()

	nested := true
	truncateAfterLines := float32(3)

	treemapSize := kbapi.KibanaHTTPAPIsLegendSizeL
	treemapVisibility := kbapi.KibanaHTTPAPIsTreemapLegendVisibilityVisible
	treemapAPI := &kbapi.KibanaHTTPAPIsTreemapLegend{
		Nested:             &nested,
		Size:               &treemapSize,
		TruncateAfterLines: &truncateAfterLines,
		Visibility:         &treemapVisibility,
	}

	mosaicSize := kbapi.KibanaHTTPAPIsLegendSizeL
	mosaicVisibility := kbapi.KibanaHTTPAPIsMosaicLegendVisibilityVisible
	mosaicAPI := &kbapi.KibanaHTTPAPIsMosaicLegend{
		Nested:             &nested,
		Size:               &mosaicSize,
		TruncateAfterLines: &truncateAfterLines,
		Visibility:         &mosaicVisibility,
	}

	treemapModel := &models.PartitionLegendModel{}
	PartitionLegendFromTreemapLegend(treemapModel, treemapAPI)

	mosaicModel := &models.PartitionLegendModel{}
	PartitionLegendFromMosaicLegend(mosaicModel, mosaicAPI)

	// Treemap and mosaic share an identical shape, so both directions must produce identical results.
	assert.Equal(t, treemapModel, mosaicModel)

	assert.Equal(t, types.BoolValue(true), treemapModel.Nested)
	assert.Equal(t, types.StringValue("l"), treemapModel.Size)
	assert.Equal(t, types.Int64Value(3), treemapModel.TruncateAfterLine)
	assert.Equal(t, types.StringValue("visible"), treemapModel.Visible)
}

func TestPartitionLegendFromTreemapAndMosaicLegend_PartiallyPopulated(t *testing.T) {
	t.Parallel()

	treemapAPI := &kbapi.KibanaHTTPAPIsTreemapLegend{}
	mosaicAPI := &kbapi.KibanaHTTPAPIsMosaicLegend{}

	treemapModel := &models.PartitionLegendModel{}
	PartitionLegendFromTreemapLegend(treemapModel, treemapAPI)

	mosaicModel := &models.PartitionLegendModel{}
	PartitionLegendFromMosaicLegend(mosaicModel, mosaicAPI)

	assert.Equal(t, treemapModel, mosaicModel)
	assert.Equal(t, types.BoolNull(), treemapModel.Nested)
	assert.Equal(t, types.StringNull(), treemapModel.Size)
	assert.Equal(t, types.Int64Null(), treemapModel.TruncateAfterLine)
	assert.Equal(t, types.StringNull(), treemapModel.Visible)
}

func TestPartitionLegendToTreemapAndMosaicLegend_FullyPopulated(t *testing.T) {
	t.Parallel()

	m := &models.PartitionLegendModel{
		Nested:            types.BoolValue(true),
		Size:              types.StringValue("s"),
		TruncateAfterLine: types.Int64Value(5),
		Visible:           types.StringValue("hidden"),
	}

	treemap := PartitionLegendToTreemapLegend(m)
	mosaic := PartitionLegendToMosaicLegend(m)

	require := assert.New(t)
	require.NotNil(treemap.Nested)
	require.True(*treemap.Nested)
	require.NotNil(treemap.Size)
	require.Equal(kbapi.KibanaHTTPAPIsLegendSizeS, *treemap.Size)
	require.NotNil(treemap.TruncateAfterLines)
	require.InDelta(float32(5), *treemap.TruncateAfterLines, 0)
	require.NotNil(treemap.Visibility)
	require.Equal(kbapi.KibanaHTTPAPIsTreemapLegendVisibilityHidden, *treemap.Visibility)

	require.NotNil(mosaic.Nested)
	require.True(*mosaic.Nested)
	require.NotNil(mosaic.Size)
	require.Equal(kbapi.KibanaHTTPAPIsLegendSizeS, *mosaic.Size)
	require.NotNil(mosaic.TruncateAfterLines)
	require.InDelta(float32(5), *mosaic.TruncateAfterLines, 0)
	require.NotNil(mosaic.Visibility)
	require.Equal(kbapi.KibanaHTTPAPIsMosaicLegendVisibilityHidden, *mosaic.Visibility)
}

func TestPartitionLegendToTreemapAndMosaicLegend_UnknownFields(t *testing.T) {
	t.Parallel()

	m := &models.PartitionLegendModel{
		Nested:            types.BoolUnknown(),
		Size:              types.StringValue("auto"),
		TruncateAfterLine: types.Int64Unknown(),
		Visible:           types.StringUnknown(),
	}

	treemap := PartitionLegendToTreemapLegend(m)
	mosaic := PartitionLegendToMosaicLegend(m)

	assert.Nil(t, treemap.Nested)
	assert.Nil(t, treemap.TruncateAfterLines)
	assert.Nil(t, treemap.Visibility)
	require := assert.New(t)
	require.NotNil(treemap.Size)
	require.Equal(kbapi.KibanaHTTPAPIsLegendSizeAuto, *treemap.Size)

	assert.Nil(t, mosaic.Nested)
	assert.Nil(t, mosaic.TruncateAfterLines)
	assert.Nil(t, mosaic.Visibility)
	require.NotNil(mosaic.Size)
	require.Equal(kbapi.KibanaHTTPAPIsLegendSizeAuto, *mosaic.Size)
}
