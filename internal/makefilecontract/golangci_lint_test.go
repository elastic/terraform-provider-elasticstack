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

package makefilecontract_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Guards REQ-041: the golangci-lint Make recipe must pass ./... so lint covers the full module.
func TestMakefileGolangCILintUsesModuleWildcard(t *testing.T) {
	t.Parallel()

	makefilePath := filepath.Join("..", "..", "Makefile")
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	const want = "golangci-lint-custom run --max-same-issues=0 $(GOLANGCIFLAGS) ./..."
	if !strings.Contains(string(data), want) {
		t.Fatalf("Makefile must contain golangci-lint invocation %q", want)
	}
}
