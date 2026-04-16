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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const memoryVersion = 1

// Memory persists Kibana spec-impact workflow state for baseline tracking and dedupe.
type Memory struct {
	Version               int                       `json:"version"`
	LastAnalyzedTargetSHA string                    `json:"last_analyzed_target_sha"`
	ReportedFingerprints  map[string]FingerprintRec `json:"reported_fingerprints"`
}

// FingerprintRec records a previously emitted impact fingerprint.
type FingerprintRec struct {
	EntityName   string    `json:"entity_name"`
	EntityType   string    `json:"entity_type"`
	BaselineSHA  string    `json:"baseline_sha"`
	TargetSHA    string    `json:"target_sha"`
	RecordedAt   time.Time `json:"recorded_at"`
	Fingerprint  string    `json:"fingerprint"`
}

func loadMemory(path string) (*Memory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read memory: %w", err)
	}
	var m Memory
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}
	if m.Version == 0 {
		m.Version = memoryVersion
	}
	if m.ReportedFingerprints == nil {
		m.ReportedFingerprints = make(map[string]FingerprintRec)
	}
	return &m, nil
}

func saveMemory(path string, m *Memory) error {
	if m.Version == 0 {
		m.Version = memoryVersion
	}
	if m.ReportedFingerprints == nil {
		m.ReportedFingerprints = make(map[string]FingerprintRec)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir for memory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".kibana-spec-impact-memory-*.json")
	if err != nil {
		return fmt.Errorf("temp memory: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write temp memory: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close temp memory: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename memory: %w", err)
	}
	return nil
}

func bootstrapMemoryFromSeed(targetPath, seedPath string) error {
	var mem *Memory
	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		mem = &Memory{Version: memoryVersion, ReportedFingerprints: map[string]FingerprintRec{}}
	} else if err != nil {
		return fmt.Errorf("stat seed: %w", err)
	} else {
		var err error
		mem, err = loadMemory(seedPath)
		if err != nil {
			return fmt.Errorf("load seed: %w", err)
		}
	}
	return saveMemory(targetPath, mem)
}

// impactFingerprint is a deterministic identity for baseline→target entity impacts.
func impactFingerprint(baselineSHA, targetSHA, entityName, entityType string, symbols []string) string {
	s := append([]string{}, symbols...)
	sort.Strings(s)
	h := sha256.New()
	fmt.Fprintf(h, "%s\n%s\n%s\n%s\n%s", baselineSHA, targetSHA, entityName, entityType, strings.Join(s, "\n"))
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

func memoryIsReported(m *Memory, fp string) bool {
	if m == nil {
		return false
	}
	_, ok := m.ReportedFingerprints[fp]
	return ok
}

func memoryRecordImpact(m *Memory, baselineSHA, targetSHA, entityName, entityType string, symbols []string) (FingerprintRec, error) {
	if m == nil {
		return FingerprintRec{}, errors.New("memory is nil")
	}
	fp := impactFingerprint(baselineSHA, targetSHA, entityName, entityType, symbols)
	rec := FingerprintRec{
		EntityName:  entityName,
		EntityType:  entityType,
		BaselineSHA: baselineSHA,
		TargetSHA:   targetSHA,
		RecordedAt:  time.Now().UTC(),
		Fingerprint: fp,
	}
	m.ReportedFingerprints[fp] = rec
	m.LastAnalyzedTargetSHA = targetSHA
	return rec, nil
}
