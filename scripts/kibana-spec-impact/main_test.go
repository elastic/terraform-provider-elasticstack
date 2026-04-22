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

import "testing"

func TestValidateMemoryRecordCLIRequiresIssuedWithImpacts(t *testing.T) {
	r := &ImpactReport{
		HighConfidence: []ImpactedEntity{{EntityName: "e1"}},
	}
	if validateMemoryRecordCLI(r, "") == nil {
		t.Fatal("expected error when impacts present and issued path empty")
	}
	if err := validateMemoryRecordCLI(r, "/tmp/will-not-be-read-here"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateMemoryRecordCLIIssuedOptionalWithoutImpacts(t *testing.T) {
	r := &ImpactReport{HighConfidence: nil}
	if err := validateMemoryRecordCLI(r, ""); err != nil {
		t.Fatal(err)
	}
}
