// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
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

package engine_test

import (
	"regexp"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/scripts/changelog/internal/engine"
)

func TestValidateMode_invalid(t *testing.T) {
	err := engine.ValidateModeAndTargetVersion("staging", "")
	if err == nil || !regexp.MustCompile(`(?i)Invalid changelog mode`).MatchString(err.Error()) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateMode_releaseMissingTarget(t *testing.T) {
	err := engine.ValidateModeAndTargetVersion(engine.ModeRelease, "")
	if err == nil || !regexp.MustCompile(`targetVersion`).MatchString(err.Error()) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateMode_releaseLeadingVRejected(t *testing.T) {
	err := engine.ValidateModeAndTargetVersion(engine.ModeRelease, "v1.0.0")
	if err == nil || !regexp.MustCompile(`targetVersion`).MatchString(err.Error()) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateMode_releaseSemverOK(t *testing.T) {
	err := engine.ValidateModeAndTargetVersion(engine.ModeRelease, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}
}
