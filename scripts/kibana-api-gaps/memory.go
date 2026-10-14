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
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/elastic/terraform-provider-elasticstack/scripts/internal/memoryio"
)

const (
	verdictFile     = "file"
	verdictSuppress = "suppress"
)

// Record is the persisted verdict for one gap group. For a `file` verdict Issue holds the group
// key, which is also embedded in the filed issue as a visible marker line.
type Record struct {
	Fingerprint string `json:"fingerprint"`
	Verdict     string `json:"verdict"`
	Reason      string `json:"reason,omitempty"`
	Date        string `json:"date,omitempty"`
	Issue       string `json:"issue,omitempty"`
}

// Memory is the content of state.json on the memory branch.
type Memory struct {
	Groups map[string]Record `json:"groups"`
	// Resources maps a resource name to the date a group of it was last triaged.
	Resources map[string]string `json:"resources"`
}

func newMemory() *Memory {
	return &Memory{Groups: map[string]Record{}, Resources: map[string]string{}}
}

// covers reports whether the group has a verdict whose fingerprint still matches; a record with
// a different fingerprint has expired.
func (m *Memory) covers(g Group) bool {
	rec, ok := m.Groups[g.Key]
	return ok && rec.Fingerprint == g.Fingerprint
}

// loadMemory reads state.json; an absent file is an empty memory.
func loadMemory(path string) (*Memory, error) {
	mem := newMemory()
	if err := memoryio.ReadJSON(path, mem); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return newMemory(), nil
		}
		return nil, err
	}
	if mem.Groups == nil {
		mem.Groups = map[string]Record{}
	}
	if mem.Resources == nil {
		mem.Resources = map[string]string{}
	}
	return mem, nil
}

// saveMemory writes state.json atomically.
func saveMemory(path string, mem *Memory) error {
	return memoryio.AtomicWriteJSON(path, mem, filepath.Base(path)+".tmp-*")
}
