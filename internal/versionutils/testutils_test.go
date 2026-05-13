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

package versionutils

import (
	"context"
	"errors"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

var (
	testVer810 = version.Must(version.NewVersion("8.10.0"))
	testVer811 = version.Must(version.NewVersion("8.11.0"))
	testVer812 = version.Must(version.NewVersion("8.12.0"))
)

func stubServerInfo(v *version.Version, buildFlavor string, fetchErr error) serverInfoGetter {
	return func(context.Context) (*version.Version, string, error) {
		return v, buildFlavor, fetchErr
	}
}

func TestCheckSkip_noFetchWhenNothingRequired(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(nil, "", errors.New("stub should not run"))

	skip, reason, err := checkSkip(context.Background(), nil, nil, FlavorAny, stub)
	require.NoError(t, err)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_belowMinimumVersion(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer810, "default", nil)

	skip, reason, err := checkSkip(context.Background(), testVer811, nil, FlavorAny, stub)
	require.NoError(t, err)
	require.True(t, skip)
	require.Contains(t, reason, "below required minimum")
}

func TestCheckSkip_atOrAboveMinimumVersion(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, "default", nil)

	skip, _, err := checkSkip(context.Background(), testVer811, nil, FlavorAny, stub)
	require.NoError(t, err)
	require.False(t, skip)
}

func TestCheckSkip_serverlessBypassesVersion(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer810, clients.ServerlessFlavor, nil)

	skip, _, err := checkSkip(context.Background(), testVer811, nil, FlavorAny, stub)
	require.NoError(t, err)
	require.False(t, skip)
}

func TestCheckSkip_flavorStatefulOnServerlessSkips(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, clients.ServerlessFlavor, nil)

	skip, reason, err := checkSkip(context.Background(), testVer811, nil, FlavorStateful, stub)
	require.NoError(t, err)
	require.True(t, skip)
	require.Contains(t, reason, "stateful")
}

func TestCheckSkip_flavorServerlessOnStatefulSkips(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, "default", nil)

	skip, reason, err := checkSkip(context.Background(), nil, nil, FlavorServerless, stub)
	require.NoError(t, err)
	require.True(t, skip)
	require.Contains(t, reason, "serverless")
}

func TestCheckSkip_fetchError(t *testing.T) {
	t.Parallel()

	fetchErr := errors.New("cannot create client")
	stub := stubServerInfo(nil, "", fetchErr)

	skip, reason, err := checkSkip(context.Background(), testVer811, nil, FlavorAny, stub)
	require.ErrorIs(t, err, fetchErr)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_constraintsStatefulSatisfiedNoSkip(t *testing.T) {
	t.Parallel()

	constraints, err := version.NewConstraint(">=8.9.0,!=8.11.0")
	require.NoError(t, err)

	stub := stubServerInfo(testVer812, "default", nil)

	skip, reason, err := checkSkip(context.Background(), nil, constraints, FlavorAny, stub)
	require.NoError(t, err)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_constraintsStatefulViolatedSkips(t *testing.T) {
	t.Parallel()

	constraints, err := version.NewConstraint(">=8.9.0,!=8.11.0")
	require.NoError(t, err)

	stub := stubServerInfo(testVer811, "default", nil)

	skip, reason, err := checkSkip(context.Background(), nil, constraints, FlavorAny, stub)
	require.NoError(t, err)
	require.True(t, skip)
	require.Contains(t, reason, "does not satisfy constraints")
}

func TestCheckSkip_constraintsServerlessBypassesEvenWhenViolated(t *testing.T) {
	t.Parallel()

	constraints, err := version.NewConstraint(">=8.9.0,!=8.11.0")
	require.NoError(t, err)

	stub := stubServerInfo(testVer811, clients.ServerlessFlavor, nil)

	skip, reason, err := checkSkip(context.Background(), nil, constraints, FlavorAny, stub)
	require.NoError(t, err)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_flavorStatefulOnStatefulNoSkip(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, "default", nil)

	skip, reason, err := checkSkip(context.Background(), nil, nil, FlavorStateful, stub)
	require.NoError(t, err)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_flavorServerlessOnServerlessNoSkip(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, clients.ServerlessFlavor, nil)

	skip, reason, err := checkSkip(context.Background(), nil, nil, FlavorServerless, stub)
	require.NoError(t, err)
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_unknownFlavorReturnsError(t *testing.T) {
	t.Parallel()

	stub := stubServerInfo(testVer812, "default", nil)

	skip, reason, err := checkSkip(context.Background(), nil, nil, Flavor(99), stub)
	require.ErrorContains(t, err, "unknown acceptance test flavor")
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestCheckSkip_getServerInfoParseErrorPropagates(t *testing.T) {
	t.Parallel()

	parseErr := errors.New("failed to parse the elasticsearch server version: invalid semver")
	stub := stubServerInfo(nil, "", parseErr)

	constraints, err := version.NewConstraint(">=8.0.0")
	require.NoError(t, err)

	skip, reason, err := checkSkip(context.Background(), nil, constraints, FlavorAny, stub)
	require.ErrorIs(t, err, parseErr)
	require.ErrorContains(t, err, "failed to parse the elasticsearch server version")
	require.False(t, skip)
	require.Empty(t, reason)
}

func TestFlavor_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		flavor Flavor
		want   string
	}{
		{name: "Any", flavor: FlavorAny, want: "Any"},
		{name: "Stateful", flavor: FlavorStateful, want: "Stateful"},
		{name: "Serverless", flavor: FlavorServerless, want: "Serverless"},
		{name: "unknown", flavor: Flavor(99), want: "Flavor(99)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.flavor.String())
		})
	}
}
