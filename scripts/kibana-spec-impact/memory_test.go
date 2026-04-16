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
// software distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, either express or implied.  See the
// License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"
)

func TestImpactFingerprintStable(t *testing.T) {
	a := impactFingerprint("b1", "t1", "elasticstack_kibana_foo", "resource", []string{"SymA", "SymB"})
	b := impactFingerprint("b1", "t1", "elasticstack_kibana_foo", "resource", []string{"SymB", "SymA"})
	if a != b {
		t.Fatalf("fingerprint should not depend on symbol order: %q vs %q", a, b)
	}
	c := impactFingerprint("b1", "t2", "elasticstack_kibana_foo", "resource", []string{"SymA", "SymB"})
	if a == c {
		t.Fatal("expected different fingerprint for different target")
	}
}

func TestMemoryRecordImpact(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	rec, err := memoryRecordImpact(m, "b", "t", "elasticstack_kibana_x", "resource", []string{"S"})
	if err != nil {
		t.Fatal(err)
	}
	if !memoryIsReported(m, rec.Fingerprint) {
		t.Fatal("expected fingerprint recorded")
	}
	if m.LastAnalyzedTargetSHA != "t" {
		t.Fatalf("last target: %q", m.LastAnalyzedTargetSHA)
	}
}
