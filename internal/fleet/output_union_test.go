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

package fleet

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func noopHandlers(t *testing.T, wantCalled string) OutputUnionHandlers {
	var got string
	record := func(name string) {
		got = name
	}
	t.Cleanup(func() {
		if wantCalled != "" && got != wantCalled {
			t.Errorf("handler called = %q, want %q", got, wantCalled)
		}
	})

	return OutputUnionHandlers{
		Elasticsearch: func(_ context.Context, _ *kbapi.KibanaHTTPAPIsOutputResponseElasticsearch) diag.Diagnostics {
			record("elasticsearch")
			return nil
		},
		Logstash: func(_ context.Context, _ *kbapi.KibanaHTTPAPIsOutputResponseLogstash) diag.Diagnostics {
			record("logstash")
			return nil
		},
		Kafka: func(_ context.Context, _ *kbapi.KibanaHTTPAPIsOutputResponseKafka) diag.Diagnostics {
			record("kafka")
			return nil
		},
		RemoteElasticsearch: func(_ context.Context, _ *kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch) diag.Diagnostics {
			record("remote_elasticsearch")
			return nil
		},
	}
}

func TestDispatchOutputUnion(t *testing.T) {
	t.Run("nil union returns no diagnostics and calls no handler", func(t *testing.T) {
		diags := DispatchOutputUnion(t.Context(), nil, noopHandlers(t, ""))
		if diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want none", diags)
		}
	})

	t.Run("elasticsearch variant dispatches to Elasticsearch handler", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.FromKibanaHTTPAPIsOutputResponseElasticsearch(kbapi.KibanaHTTPAPIsOutputResponseElasticsearch{
			Name: "es-output",
			Type: kbapi.KibanaHTTPAPIsOutputResponseElasticsearchTypeElasticsearch,
		}); err != nil {
			t.Fatalf("FromKibanaHTTPAPIsOutputResponseElasticsearch() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, "elasticsearch"))
		if diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want none", diags)
		}
	})

	t.Run("logstash variant dispatches to Logstash handler", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.FromKibanaHTTPAPIsOutputResponseLogstash(kbapi.KibanaHTTPAPIsOutputResponseLogstash{
			Name: "logstash-output",
			Type: kbapi.KibanaHTTPAPIsOutputResponseLogstashTypeLogstash,
		}); err != nil {
			t.Fatalf("FromKibanaHTTPAPIsOutputResponseLogstash() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, "logstash"))
		if diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want none", diags)
		}
	})

	t.Run("kafka variant dispatches to Kafka handler", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.FromKibanaHTTPAPIsOutputResponseKafka(kbapi.KibanaHTTPAPIsOutputResponseKafka{
			Name:     "kafka-output",
			Type:     kbapi.KibanaHTTPAPIsOutputResponseKafkaTypeKafka,
			AuthType: kbapi.KibanaHTTPAPIsOutputResponseKafkaAuthTypeNone,
		}); err != nil {
			t.Fatalf("FromKibanaHTTPAPIsOutputResponseKafka() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, "kafka"))
		if diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want none", diags)
		}
	})

	t.Run("remote_elasticsearch variant dispatches to RemoteElasticsearch handler", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.FromKibanaHTTPAPIsOutputResponseRemoteElasticsearch(kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch{
			Name: "remote-es-output",
			Type: kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearchTypeRemoteElasticsearch,
		}); err != nil {
			t.Fatalf("FromKibanaHTTPAPIsOutputResponseRemoteElasticsearch() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, "remote_elasticsearch"))
		if diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want none", diags)
		}
	})

	t.Run("unresolvable discriminator adds an error diagnostic and calls no handler", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.UnmarshalJSON([]byte(`{"type":"not_a_real_output_type"}`)); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, ""))
		if !diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want an error", diags)
		}
	})

	t.Run("resolved variant with no matching case adds an error diagnostic", func(t *testing.T) {
		var union kbapi.OutputUnion
		if err := union.UnmarshalJSON([]byte(`{"type":"otlp"}`)); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		diags := DispatchOutputUnion(t.Context(), &union, noopHandlers(t, ""))
		if !diags.HasError() {
			t.Errorf("DispatchOutputUnion() diags = %v, want an error", diags)
		}
	})
}
