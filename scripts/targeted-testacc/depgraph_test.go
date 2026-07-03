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
