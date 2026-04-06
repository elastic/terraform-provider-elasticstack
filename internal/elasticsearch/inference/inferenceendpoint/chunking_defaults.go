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

package inferenceendpoint

import (
	"encoding/json"
	"maps"
)

const chunkingStrategySentence = "sentence"

// populateChunkingSettingsDefaults fills documented Elasticsearch defaults for
// chunking_settings so plan/state semantic equality matches when the API echoes
// defaults the user did not set. Defaults follow the inference API docs
// (e.g. strategy sentence → max_chunk_size 250, sentence_overlap 1).
func populateChunkingSettingsDefaults(model map[string]any) map[string]any {
	out := deepCloneJSONMap(model)
	if out == nil {
		out = make(map[string]any)
	}

	strategy, _ := stringFromJSON(out["strategy"])
	if strategy == "" {
		out["strategy"] = chunkingStrategySentence
		strategy = chunkingStrategySentence
	}

	switch strategy {
	case chunkingStrategySentence:
		if _, ok := out["max_chunk_size"]; !ok {
			out["max_chunk_size"] = float64(250)
		}
		if _, ok := out["sentence_overlap"]; !ok {
			out["sentence_overlap"] = float64(1)
		}
	case "word":
		if _, ok := out["max_chunk_size"]; !ok {
			out["max_chunk_size"] = float64(250)
		}
		if _, ok := out["overlap"]; !ok {
			out["overlap"] = float64(100)
		}
	case "none", "recursive":
		// No single set of defaults applies; user-supplied shape is preserved.
	}

	return out
}

func deepCloneJSONMap(model map[string]any) map[string]any {
	if model == nil {
		return nil
	}
	b, err := json.Marshal(model)
	if err != nil {
		return shallowCopyJSONMap(model)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return shallowCopyJSONMap(model)
	}
	return out
}

func shallowCopyJSONMap(model map[string]any) map[string]any {
	out := make(map[string]any, len(model))
	maps.Copy(out, model)
	return out
}

func stringFromJSON(v any) (string, bool) {
	switch s := v.(type) {
	case string:
		return s, true
	default:
		return "", false
	}
}
