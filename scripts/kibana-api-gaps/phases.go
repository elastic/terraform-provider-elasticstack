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
	"fmt"
	"io"
	"time"
)

// phaseLog writes one stderr line as each discovery phase starts and ends, so a CI step that
// hangs or times out shows which phase it was in. A nil *phaseLog is a no-op.
type phaseLog struct {
	w     io.Writer
	start time.Time
	last  time.Time
}

func newPhaseLog(w io.Writer) *phaseLog {
	now := time.Now()
	return &phaseLog{w: w, start: now, last: now}
}

// begin announces a phase that is about to run.
func (p *phaseLog) begin(name string) {
	if p == nil {
		return
	}
	fmt.Fprintf(p.w, "discover: %s: started (elapsed %s)\n", name, time.Since(p.start).Round(time.Millisecond))
	p.last = time.Now()
}

// end reports how long the phase announced by the last begin took.
func (p *phaseLog) end(name string) {
	if p == nil {
		return
	}
	fmt.Fprintf(p.w, "discover: %s: done in %s (elapsed %s)\n", name,
		time.Since(p.last).Round(time.Millisecond), time.Since(p.start).Round(time.Millisecond))
}

// report prints a phase whose duration was accumulated elsewhere.
func (p *phaseLog) report(name string, d time.Duration) {
	if p == nil {
		return
	}
	fmt.Fprintf(p.w, "discover: %s: %s\n", name, d.Round(time.Millisecond))
}

func (p *phaseLog) total() {
	if p == nil {
		return
	}
	fmt.Fprintf(p.w, "discover: total: %s\n", time.Since(p.start).Round(time.Millisecond))
}
