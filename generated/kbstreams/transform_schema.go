//go:build ignore
// +build ignore

// TEMPORARY (POC): Minimal transformer that trims a full Kibana OAS down to the
// Streams-related paths only. This keeps oapi-codegen focused on the Streams
// surface while avoiding issues in unrelated APIs (attack discovery, cases,
// fleet settings, etc.).
//
// This program is intentionally small and self-contained so that it can be
// iterated on or replaced once Streams is fully integrated into the main
// Kibana OAS pipeline.
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	inFile := flag.String("i", "", "input file")
	outFile := flag.String("o", "", "output file")
	flag.Parse()

	if *inFile == "" || *outFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	data, err := os.ReadFile(*inFile)
	if err != nil {
		log.Fatalf("failed to read input schema %q: %v", *inFile, err)
	}

	var root map[string]any
	if err := yaml.Unmarshal(data, &root); err != nil {
		log.Fatalf("failed to unmarshal input schema %q: %v", *inFile, err)
	}

	pathsAny, ok := root["paths"]
	if !ok {
		log.Fatalf("input schema has no top-level \"paths\" key")
	}

	pathsMap, ok := pathsAny.(map[string]any)
	if !ok {
		log.Fatalf("input schema \"paths\" is not an object")
	}

	streamsPaths := map[string]any{}
	for p, v := range pathsMap {
		if strings.HasPrefix(p, "/api/streams") {
			streamsPaths[p] = v
		}
	}

	if len(streamsPaths) == 0 {
		log.Fatalf("no /api/streams* paths found in input schema %q", *inFile)
	}

	root["paths"] = streamsPaths

	out, err := yaml.Marshal(root)
	if err != nil {
		log.Fatalf("failed to marshal filtered schema %q: %v", *outFile, err)
	}

	if err := os.WriteFile(*outFile, out, 0o664); err != nil {
		log.Fatalf("failed to write filtered schema %q: %v", *outFile, err)
	}
}
