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

variable "job_id" {
  description = "The ML job ID"
  type        = string
}

variable "datafeed_id" {
  description = "The ML datafeed ID"
  type        = string
}

variable "endpoints" {
  type = list(string)
}

variable "username" {
  type    = string
  default = ""
}

variable "password" {
  type    = string
  default = ""
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = "Test job for datafeed explicit connection"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = var.username
    password  = var.password
    insecure  = true
  }
}

resource "elasticstack_elasticsearch_ml_datafeed" "test" {
  datafeed_id = var.datafeed_id
  job_id      = elasticstack_elasticsearch_ml_anomaly_detection_job.test.job_id
  indices     = ["test-index-*", "test-index-2-*"] # Added index to exercise update path

  elasticsearch_connection {
    endpoints = var.endpoints
    username  = var.username
    password  = var.password
    insecure  = true
  }

  depends_on = [elasticstack_elasticsearch_ml_anomaly_detection_job.test]
}
