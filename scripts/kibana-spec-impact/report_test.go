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
	"time"
)

func TestImpactEntryForEntitySuppressesDuplicate(t *testing.T) {
	fp := impactFingerprint("b", "t", "elasticstack_kibana_x", "resource", []string{"Sym"})
	mem := &Memory{
		Version:              1,
		ReportedFingerprints: map[string]FingerprintRec{fp: {Fingerprint: fp, RecordedAt: time.Now().UTC()}},
	}
	e := Entity{Type: "resource", Name: "elasticstack_kibana_x", PkgPath: "p"}
	high, sup := impactEntryForEntity(mem, "b", "t", e, []string{"Sym"})
	if high != nil || sup == nil {
		t.Fatalf("expected suppressed, got high=%v sup=%v", high, sup)
	}
	if sup.EntityName != e.Name || sup.Reason != "duplicate_fingerprint" {
		t.Fatalf("unexpected suppression: %+v", sup)
	}
}

func TestImpactEntryForEntityHighConfidence(t *testing.T) {
	mem := &Memory{Version: 1, ReportedFingerprints: map[string]FingerprintRec{}}
	e := Entity{Type: "resource", Name: "elasticstack_kibana_x", PkgPath: "p"}
	high, sup := impactEntryForEntity(mem, "b", "t", e, []string{"Sym"})
	if sup != nil || high == nil {
		t.Fatalf("expected high confidence, got high=%v sup=%v", high, sup)
	}
	if high.EntityName != e.Name || high.Confidence != "high" {
		t.Fatalf("unexpected entry: %+v", high)
	}
}
