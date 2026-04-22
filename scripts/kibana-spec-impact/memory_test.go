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

func TestMemoryAddFingerprintDoesNotAdvanceBaseline(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	rec, err := memoryAddFingerprint(m, "b", "t", "elasticstack_kibana_x", "resource", []string{"S"})
	if err != nil {
		t.Fatal(err)
	}
	if !memoryIsReported(m, rec.Fingerprint) {
		t.Fatal("expected fingerprint recorded")
	}
	if m.LastAnalyzedTargetSHA != "" {
		t.Fatalf("baseline should not advance from fingerprint alone: %q", m.LastAnalyzedTargetSHA)
	}
}

func TestAdvanceMemoryBaseline(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	advanceMemoryBaseline(m, "abc123")
	if m.LastAnalyzedTargetSHA != "abc123" {
		t.Fatalf("got %q", m.LastAnalyzedTargetSHA)
	}
}

func TestRecordIssuedFingerprintsPartial(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	report := &ImpactReport{
		BaselineSHA: "b",
		TargetSHA:   "t",
		HighConfidence: []ImpactedEntity{
			{EntityName: "e1", EntityType: "resource", MatchedSymbols: []string{"A"}},
			{EntityName: "e2", EntityType: "resource", MatchedSymbols: []string{"B"}},
		},
	}
	n, err := recordIssuedFingerprints(m, report, []string{"e1"})
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	fp1 := impactFingerprint("b", "t", "e1", "resource", []string{"A"})
	if !memoryIsReported(m, fp1) {
		t.Fatal("e1 not recorded")
	}
	fp2 := impactFingerprint("b", "t", "e2", "resource", []string{"B"})
	if memoryIsReported(m, fp2) {
		t.Fatal("e2 should not be recorded when not issued")
	}
}

func TestRecordIssuedFingerprintsUnknownEntityErrors(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	report := &ImpactReport{
		BaselineSHA:    "b",
		TargetSHA:      "t",
		HighConfidence: []ImpactedEntity{{EntityName: "e1", EntityType: "resource", MatchedSymbols: []string{"A"}}},
	}
	_, err := recordIssuedFingerprints(m, report, []string{"missing"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRecordIssuedFingerprintsDeduplicatesNames(t *testing.T) {
	m := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	report := &ImpactReport{
		BaselineSHA:    "b",
		TargetSHA:      "t",
		HighConfidence: []ImpactedEntity{{EntityName: "e1", EntityType: "resource", MatchedSymbols: []string{"A"}}},
	}
	n, err := recordIssuedFingerprints(m, report, []string{"e1", "e1"})
	if err != nil || n != 1 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}
