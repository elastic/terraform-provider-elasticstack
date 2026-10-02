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

package output

import (
	"reflect"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
)

// assertNoZeroFields fails the test if any field of v (a struct value) is
// still the zero value for its type. This guards against buildCommonOutputReadData
// or populateCommonOutputFields silently leaving a field unwired (e.g. a typo
// in outputReadFieldNameOverrides, or a future kbapi regeneration renaming a
// field) when every field of the fully-populated fixture used by the test
// is expected to have been copied across.
func assertNoZeroFields(t *testing.T, v any) {
	t.Helper()

	rv := reflect.ValueOf(v)
	rt := rv.Type()
	for i := range rt.NumField() {
		field := rv.Field(i)
		zero := reflect.Zero(field.Type())
		assert.Falsef(t, reflect.DeepEqual(field.Interface(), zero.Interface()),
			"field %q: expected to be wired from the source struct, but it is zero-valued", rt.Field(i).Name)
	}
}

// TestBuildCommonOutputReadData locks in that buildCommonOutputReadData
// derives every commonOutputReadData field from the matching field of the
// generated per-output-type kbapi response struct, independent of the
// per-type fromAPI*Model wrapper functions that call it.
func TestBuildCommonOutputReadData(t *testing.T) {
	t.Parallel()

	hosts := []string{"https://127.0.0.1:9200"}
	id := "output-id"
	caSha256 := "ca-sha-256"
	caTrustedFingerprint := "ca-fingerprint"
	isDefault := true
	isDefaultMonitoring := true
	configYaml := "a: 1\n"
	ssl := &kbapi.KibanaHTTPAPIsOutputResponseSsl{}

	t.Run("Elasticsearch", func(t *testing.T) {
		data := &kbapi.KibanaHTTPAPIsOutputResponseElasticsearch{
			Id:                   &id,
			Name:                 "es-output",
			Type:                 kbapi.KibanaHTTPAPIsOutputResponseElasticsearchTypeElasticsearch,
			Hosts:                hosts,
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			ConfigYaml:           &configYaml,
			Ssl:                  ssl,
		}

		got := buildCommonOutputReadData(data)
		assertNoZeroFields(t, got)
		assert.Equal(t, "es-output", got.Name)
		assert.Equal(t, "elasticsearch", got.OutputType)
		assert.Same(t, ssl, got.Ssl)
	})

	t.Run("Logstash", func(t *testing.T) {
		data := &kbapi.KibanaHTTPAPIsOutputResponseLogstash{
			Id:                   &id,
			Name:                 "logstash-output",
			Type:                 kbapi.KibanaHTTPAPIsOutputResponseLogstashTypeLogstash,
			Hosts:                hosts,
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			ConfigYaml:           &configYaml,
			Ssl:                  ssl,
		}

		got := buildCommonOutputReadData(data)
		assertNoZeroFields(t, got)
		assert.Equal(t, "logstash-output", got.Name)
		assert.Equal(t, "logstash", got.OutputType)
		assert.Same(t, ssl, got.Ssl)
	})

	t.Run("Kafka", func(t *testing.T) {
		data := &kbapi.KibanaHTTPAPIsOutputResponseKafka{
			Id:                   &id,
			Name:                 "kafka-output",
			Type:                 kbapi.KibanaHTTPAPIsOutputResponseKafkaTypeKafka,
			AuthType:             kbapi.KibanaHTTPAPIsOutputResponseKafkaAuthTypeNone,
			Hosts:                hosts,
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			ConfigYaml:           &configYaml,
			Ssl:                  ssl,
		}

		got := buildCommonOutputReadData(data)
		assertNoZeroFields(t, got)
		assert.Equal(t, "kafka-output", got.Name)
		assert.Equal(t, "kafka", got.OutputType)
		assert.Same(t, ssl, got.Ssl)
	})

	t.Run("RemoteElasticsearch", func(t *testing.T) {
		data := &kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearch{
			Id:                   &id,
			Name:                 "remote-es-output",
			Type:                 kbapi.KibanaHTTPAPIsOutputResponseRemoteElasticsearchTypeRemoteElasticsearch,
			Hosts:                hosts,
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			ConfigYaml:           &configYaml,
			Ssl:                  ssl,
		}

		got := buildCommonOutputReadData(data)
		assertNoZeroFields(t, got)
		assert.Equal(t, "remote-es-output", got.Name)
		assert.Equal(t, "remote_elasticsearch", got.OutputType)
		assert.Same(t, ssl, got.Ssl)
	})
}

// TestPopulateCommonOutputFields locks in that populateCommonOutputFields
// copies every commonNewOutputBody/commonUpdateOutputBody field onto the
// matching field of the generated per-output-type kbapi New/Update struct
// (ID mapping to Id), independent of the per-type toAPICreate*Model/
// toAPIUpdate*Model wrapper functions that call it.
func TestPopulateCommonOutputFields(t *testing.T) {
	t.Parallel()

	hosts := []string{"https://127.0.0.1:9200"}
	id := "output-id"
	caSha256 := "ca-sha-256"
	caTrustedFingerprint := "ca-fingerprint"
	isDefault := true
	isDefaultMonitoring := true
	configYaml := "a: 1\n"
	ssl := &kbapi.KibanaHTTPAPIsOutputSsl{}

	t.Run("NewOutputElasticsearch", func(t *testing.T) {
		f := commonNewOutputBody{
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			ConfigYaml:           &configYaml,
			Hosts:                hosts,
			ID:                   &id,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			Name:                 "es-output",
			Ssl:                  ssl,
		}

		body := kbapi.KibanaHTTPAPIsNewOutputElasticsearch{
			Type: kbapi.KibanaHTTPAPIsNewOutputElasticsearchTypeElasticsearch,
		}
		populateCommonOutputFields(&body, f)

		assert.Equal(t, &id, body.Id)
		assert.Equal(t, "es-output", body.Name)
		assert.Same(t, ssl, body.Ssl)
		assert.Equal(t, hosts, body.Hosts)
		assert.Equal(t, &caSha256, body.CaSha256)
		assert.Equal(t, &caTrustedFingerprint, body.CaTrustedFingerprint)
		assert.Equal(t, &isDefault, body.IsDefault)
		assert.Equal(t, &isDefaultMonitoring, body.IsDefaultMonitoring)
		assert.Equal(t, &configYaml, body.ConfigYaml)
		// The Type discriminator is caller-set, not part of the common body.
		assert.Equal(t, kbapi.KibanaHTTPAPIsNewOutputElasticsearchTypeElasticsearch, body.Type)
	})

	t.Run("UpdateOutputLogstash", func(t *testing.T) {
		f := commonUpdateOutputBody{
			CaSha256:             &caSha256,
			CaTrustedFingerprint: &caTrustedFingerprint,
			ConfigYaml:           &configYaml,
			Hosts:                &hosts,
			IsDefault:            &isDefault,
			IsDefaultMonitoring:  &isDefaultMonitoring,
			Name:                 &id, // reuse a *string fixture; value itself is irrelevant here
			Ssl:                  ssl,
		}

		outputType := kbapi.Logstash
		body := kbapi.KibanaHTTPAPIsUpdateOutputLogstash{
			Type: &outputType,
		}
		populateCommonOutputFields(&body, f)

		assert.Same(t, ssl, body.Ssl)
		assert.Equal(t, &hosts, body.Hosts)
		assert.Equal(t, &caSha256, body.CaSha256)
		assert.Equal(t, &caTrustedFingerprint, body.CaTrustedFingerprint)
		assert.Equal(t, &isDefault, body.IsDefault)
		assert.Equal(t, &isDefaultMonitoring, body.IsDefaultMonitoring)
		assert.Equal(t, &configYaml, body.ConfigYaml)
		assert.Equal(t, &id, body.Name)
		// Update structs have no ID field; populateCommonOutputFields must not panic
		// just because commonUpdateOutputBody has no ID field to look for.
		assert.Nil(t, body.Id)
		assert.Equal(t, &outputType, body.Type)
	})
}
