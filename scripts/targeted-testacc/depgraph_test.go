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
	"reflect"
	"testing"
)

func TestBuildReverseDepGraph(t *testing.T) {
	forward := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {},
	}

	got := BuildReverseDepGraph(forward)
	want := map[string][]string{
		"B": {"A"},
		"C": {"B"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildReverseDepGraph = %v, want %v", got, want)
	}
}

func TestBuildReverseDepGraph_DeduplicatesImporters(t *testing.T) {
	forward := map[string][]string{
		"A": {"C"},
		"B": {"C"},
		"C": {},
	}

	got := BuildReverseDepGraph(forward)
	want := map[string][]string{
		"C": {"A", "B"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BuildReverseDepGraph = %v, want %v", got, want)
	}
}

func TestWalkReverseDeps_LinearChain(t *testing.T) {
	reverse := map[string][]string{
		"B": {"A"},
		"C": {"B"},
	}

	got := WalkReverseDeps(reverse, []string{"C"})
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkReverseDeps = %v, want %v", got, want)
	}
}

func TestWalkReverseDeps_Diamond(t *testing.T) {
	forward := map[string][]string{
		"top":    {"left", "right"},
		"left":   {"bottom"},
		"right":  {"bottom"},
		"bottom": {},
	}
	reverse := BuildReverseDepGraph(forward)

	got := WalkReverseDeps(reverse, []string{"bottom"})
	want := []string{"bottom", "left", "right", "top"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkReverseDeps = %v, want %v", got, want)
	}
}

func TestWalkReverseDeps_MultipleStarts(t *testing.T) {
	reverse := map[string][]string{
		"C": {"A", "B"},
	}

	got := WalkReverseDeps(reverse, []string{"B", "C"})
	want := []string{"A", "B", "C"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkReverseDeps = %v, want %v", got, want)
	}
}

func TestWalkReverseDeps_NoImporters(t *testing.T) {
	reverse := map[string][]string{}

	got := WalkReverseDeps(reverse, []string{"leaf"})
	want := []string{"leaf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkReverseDeps = %v, want %v", got, want)
	}
}
