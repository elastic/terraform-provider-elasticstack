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

package advancedsettings

import (
	"cmp"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/stretchr/testify/require"
)

// fakeKibana serves the spaces API and the advanced settings routes, and
// records every request it receives.
type fakeKibana struct {
	mu       sync.Mutex
	spaces   map[string]bool
	settings map[string]map[string]json.RawMessage
	requests []string
	changes  []map[string]json.RawMessage
	// rejection, when set, is the message of the error response returned for every settings write.
	rejection string
	// rejectionStatus is the HTTP status of the rejection response, 400 when zero.
	rejectionStatus int
	// failReads makes every settings read return a 500 response.
	failReads bool
	// failSpaceLookups makes every spaces API lookup return a 500 response.
	failSpaceLookups bool
	// missingSettingsRoute makes every request to the settings routes return a 404 response.
	missingSettingsRoute bool
	// dropSettingsWrites closes the connection without a response to every settings write.
	dropSettingsWrites bool
}

func newFakeKibana(t *testing.T, existingSpaces ...string) (*fakeKibana, *clients.KibanaScopedClient) {
	t.Helper()
	fake := &fakeKibana{spaces: map[string]bool{}, settings: map[string]map[string]json.RawMessage{}}
	for _, space := range existingSpaces {
		fake.spaces[space] = true
	}
	srv := httptest.NewServer(http.HandlerFunc(fake.serve))
	t.Cleanup(srv.Close)
	api, err := kbapi.NewClientWithResponses(srv.URL)
	require.NoError(t, err)
	return fake, clients.NewKibanaScopedClientForTest(&kibanaoapi.Client{URL: srv.URL, HTTP: srv.Client(), API: api})
}

func (f *fakeKibana) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, r.Method+" "+r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	if spaceID, isSpaceLookup := strings.CutPrefix(r.URL.Path, "/api/spaces/space/"); isSpaceLookup {
		if f.failSpaceLookups {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"statusCode":500,"error":"Internal Server Error","message":"boom"}`))
			return
		}
		if !f.spaces[spaceID] {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"statusCode":404,"error":"Not Found","message":"Saved object [space/` + spaceID + `] not found"}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"` + spaceID + `","name":"` + spaceID + `"}`))
		return
	}

	spaceID := "default"
	if rest, scoped := strings.CutPrefix(r.URL.Path, "/s/"); scoped {
		spaceID, _, _ = strings.Cut(rest, "/")
	}
	if f.missingSettingsRoute {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"statusCode":404,"error":"Not Found","message":"Not Found"}`))
		return
	}
	if r.Method == http.MethodPost {
		if f.dropSettingsWrites {
			panic(http.ErrAbortHandler)
		}
		if f.rejection != "" {
			status := cmp.Or(f.rejectionStatus, http.StatusBadRequest)
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"statusCode": status, "error": http.StatusText(status), "message": f.rejection})
			return
		}
		var body struct {
			Changes map[string]json.RawMessage `json:"changes"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		f.changes = append(f.changes, body.Changes)
		_, _ = w.Write([]byte(`{"settings":{}}`))
		return
	}

	if f.failReads {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"statusCode":500,"error":"Internal Server Error","message":"boom"}`))
		return
	}

	entries := map[string]map[string]json.RawMessage{}
	for key, value := range f.settings[spaceID] {
		entries[key] = map[string]json.RawMessage{"userValue": value}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"settings": entries})
}

func (f *fakeKibana) recordedRequests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.requests...)
}

func (f *fakeKibana) appliedChanges() []map[string]json.RawMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]json.RawMessage(nil), f.changes...)
}
