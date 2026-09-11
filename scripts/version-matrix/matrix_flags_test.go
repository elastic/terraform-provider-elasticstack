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
)

func TestFlagsForVersion_twoDigitMinorsDoNotMatch81Ranges(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"8.10.4", "8.11.4"} {
		got := FlagsForVersion(version)
		assert.Equal(t, "ubuntu-latest", got.Runner, version)
		assert.Equal(t, "docker.elastic.co/elastic-agent/elastic-agent", got.FleetImage, version)
		assert.False(t, got.PrePullFleet, version)
	}
}

func TestFlagsForVersion_forceSyntheticsSurvivesPatchBump(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"8.14.3", "8.14.99"} {
		got := FlagsForVersion(version)
		assert.True(t, got.ForceSynthetics, version)
	}
}

func TestFlagsForVersion_forceSyntheticsMatches815(t *testing.T) {
	t.Parallel()

	assert.True(t, FlagsForVersion("8.15.5").ForceSynthetics)
}

func TestFlagsForVersion_forceSyntheticsMatches816(t *testing.T) {
	t.Parallel()

	assert.True(t, FlagsForVersion("8.16.6").ForceSynthetics)
}

func TestFlagsForVersion_forceSyntheticsMatches817(t *testing.T) {
	t.Parallel()

	assert.True(t, FlagsForVersion("8.17.10").ForceSynthetics)
}

func TestFlagsForVersion_forceSyntheticsFalseOutside814To817(t *testing.T) {
	t.Parallel()

	cases := []string{"8.13.4", "8.13.0-SNAPSHOT", "8.18.8", "8.18.0-SNAPSHOT"}
	for _, version := range cases {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			got := FlagsForVersion(version)
			assert.False(t, got.ForceSynthetics, version)
			assert.Equal(t, "ubuntu-latest", got.Runner, version)
			assert.Equal(t, "docker.elastic.co/elastic-agent/elastic-agent", got.FleetImage, version)
			assert.False(t, got.PrePullFleet, version)
		})
	}
}

func TestFlagsForVersion_80UsesUbuntu2204AndDockerHubFleet(t *testing.T) {
	t.Parallel()

	got := FlagsForVersion("8.0.1")
	assert.Equal(t, "ubuntu-22.04", got.Runner)
	assert.Equal(t, "elastic/elastic-agent", got.FleetImage)
	assert.True(t, got.PrePullFleet)
}

func TestFlagsForVersion_81UsesUbuntu2204AndDockerHubFleet(t *testing.T) {
	t.Parallel()

	got := FlagsForVersion("8.1.3")
	assert.Equal(t, "ubuntu-22.04", got.Runner)
	assert.Equal(t, "elastic/elastic-agent", got.FleetImage)
	assert.True(t, got.PrePullFleet)
}

func TestFlagsForVersion_84UsesUbuntu2204ButNotDockerHubFleet(t *testing.T) {
	t.Parallel()

	got := FlagsForVersion("8.4.3")
	assert.Equal(t, "ubuntu-22.04", got.Runner)
	assert.Equal(t, "docker.elastic.co/elastic-agent/elastic-agent", got.FleetImage)
	assert.False(t, got.PrePullFleet)
}

func TestFlagsForVersion_82And83UseUbuntu2204WithoutDockerHubFleet(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"8.2.3", "8.3.3"} {
		got := FlagsForVersion(version)
		assert.Equal(t, "ubuntu-22.04", got.Runner, version)
		assert.Equal(t, "docker.elastic.co/elastic-agent/elastic-agent", got.FleetImage, version)
		assert.False(t, got.PrePullFleet, version)
	}
}

func TestFlagsForVersion_snapshotAndGAShareMinorRangeFlags(t *testing.T) {
	t.Parallel()

	assert.Equal(t, FlagsForVersion("8.14.0"), FlagsForVersion("8.14.0-SNAPSHOT"))
}

func TestLoadMatrix_returnsPinnedVersionsAndDerivedFlags(t *testing.T) {
	t.Parallel()

	versions := []string{"8.1.3", "8.10.4", "8.14.3"}
	gotVersions, flags := LoadMatrix(versions)
	assert.Equal(t, versions, gotVersions)
	assert.Equal(t, FlagsForVersion("8.1.3"), flags["8.1.3"])
	assert.Equal(t, FlagsForVersion("8.10.4"), flags["8.10.4"])
	assert.Equal(t, FlagsForVersion("8.14.3"), flags["8.14.3"])
}
