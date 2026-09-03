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

// Fleet output type discriminators reused by the schema's `type` validator
// and by the model converter that dispatches per-output-type API calls.
const (
	outputTypeElasticsearch       = "elasticsearch"
	outputTypeLogstash            = "logstash"
	outputTypeKafka               = "kafka"
	outputTypeRemoteElasticsearch = "remote_elasticsearch"
)

// Terraform schema attribute keys shared by the output schema and the
// type-aware attribute lookup helpers.
const (
	attrName                   = "name"
	attrType                   = "type"
	attrHosts                  = "hosts"
	attrSSL                    = "ssl"
	attrCertificateAuthorities = "certificate_authorities"
	attrCertificate            = "certificate"
	attrKey                    = "key"
	attrVerificationMode       = "verification_mode"
	attrKafka                  = "kafka"
	attrValue                  = "value"
	attrHash                   = "hash"
	attrRandom                 = "random"
	attrGroupEvents            = "group_events"
	attrMechanism              = "mechanism"
)

const kafkaCompressionGzip = "gzip"
