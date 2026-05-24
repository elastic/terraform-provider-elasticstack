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

package privatelocation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func geo(lat, lon float64) *tfGeoConfigV0 {
	return &tfGeoConfigV0{
		Lat: NewFloat32PrecisionValue(lat),
		Lon: NewFloat32PrecisionValue(lon),
	}
}

func TestPreserveGeoFromInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name       string
		input      *tfGeoConfigV0
		api        *tfGeoConfigV0
		want       *tfGeoConfigV0
		wantSameAs string // "input", "api", or empty for nil
	}{
		{
			name:  "both nil",
			input: nil,
			api:   nil,
			want:  nil,
		},
		{
			name:  "input non-nil api nil",
			input: geo(42.42, 10.5),
			api:   nil,
			want:  nil,
		},
		{
			name:       "input nil api non-nil",
			input:      nil,
			api:        geo(48.8566, 2.3522),
			wantSameAs: "api",
		},
		{
			name:       "both semantically equal",
			input:      geo(42.42, 10.5),
			api:        geo(42.41999816894531, float64(float32(10.5))),
			wantSameAs: "input",
		},
		{
			name:       "lat equal lon different",
			input:      geo(42.42, 10.5),
			api:        geo(42.41999816894531, 11.0),
			wantSameAs: "api",
		},
		{
			name:       "lon equal lat different",
			input:      geo(42.42, 10.5),
			api:        geo(43.0, float64(float32(10.5))),
			wantSameAs: "api",
		},
		{
			name:       "both different",
			input:      geo(42.42, 10.5),
			api:        geo(43.0, 11.0),
			wantSameAs: "api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := preserveGeoFromInput(ctx, tt.input, tt.api)
			if tt.want == nil && tt.wantSameAs == "" {
				require.Nil(t, got)
				return
			}

			switch tt.wantSameAs {
			case "input":
				require.Same(t, tt.input, got)
				require.InDelta(t, tt.input.Lat.ValueFloat64(), got.Lat.ValueFloat64(), 0)
				require.InDelta(t, tt.input.Lon.ValueFloat64(), got.Lon.ValueFloat64(), 0)
			case "api":
				require.Same(t, tt.api, got)
			default:
				t.Fatalf("unexpected test case configuration: %+v", tt)
			}
		})
	}
}
