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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func alwaysResolve(_ string) (bool, error) { return true, nil }

func neverResolve(_ string) (bool, error) { return false, nil }

func TestComputeDesiredIncludesNewPatchWhenImagesResolve(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v8.19.21"},
		"9.6.0-SNAPSHOT",
		[]string{"8.19.17"},
		alwaysResolve,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.19.21", "9.6.0-SNAPSHOT"}, got)
}

func TestComputeDesiredRetainsPreviousGAWhenNewPatchImagesMissing(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v8.19.21"},
		"9.6.0-SNAPSHOT",
		[]string{"8.19.17"},
		neverResolve,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.19.17", "9.6.0-SNAPSHOT"}, got)
}

func TestComputeDesiredRetainsSnapshotWhenPromotionImagesMissing(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v9.6.0"},
		"9.7.0-SNAPSHOT",
		[]string{"9.6.0-SNAPSHOT"},
		neverResolve,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"9.6.0-SNAPSHOT"}, got)
}

func TestComputeDesiredOmitsBrandNewMinorWhenImagesMissing(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v8.19.21", "v9.6.0"},
		"9.7.0-SNAPSHOT",
		[]string{"8.19.21"},
		func(v string) (bool, error) { return v == "8.19.21", nil },
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.19.21", "9.7.0-SNAPSHOT"}, got)
}

func TestComputeDesiredOmitsNewTwoDigitMinorWhenUnrelatedOneDigitPinExists(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v8.1.3", "v8.10.4"},
		"9.6.0-SNAPSHOT",
		[]string{"8.1.3"},
		func(v string) (bool, error) { return v == "8.1.3", nil },
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.1.3", "9.6.0-SNAPSHOT"}, got)
}

func TestComputeDesiredPromotesSnapshotWhenGAImagesResolve(t *testing.T) {
	t.Parallel()

	got, err := ComputeDesired(
		[]string{"v9.6.0"},
		"9.7.0-SNAPSHOT",
		[]string{"9.6.0-SNAPSHOT"},
		alwaysResolve,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"9.6.0", "9.7.0-SNAPSHOT"}, got)
}
