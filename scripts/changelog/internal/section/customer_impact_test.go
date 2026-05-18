// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package section

import (
	"encoding/json"
	"testing"
)

func TestRequiresSummary(t *testing.T) {
	if !ImpactFix.RequiresSummary() || !ImpactEnhancement.RequiresSummary() || !ImpactBreaking.RequiresSummary() {
		t.Fatal()
	}
	if ImpactNone.RequiresSummary() {
		t.Fatal()
	}
}

func TestCustomerImpactMarshalJSON_roundTrip(t *testing.T) {
	for _, want := range []CustomerImpact{ImpactNone, ImpactFix, ImpactEnhancement, ImpactBreaking} {
		b, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		var got CustomerImpact
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%v != %v", got, want)
		}
	}
}

func TestCustomerImpactString(t *testing.T) {
	if ImpactBreaking.String() != "breaking" {
		t.Fatal()
	}
}
