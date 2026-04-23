## 1. Schema

- [x] 1.1 Add `verification_mode` optional string attribute to the `ssl` block in `schema.go` with a `stringvalidator.OneOf("certificate", "full", "none", "strict")` validator

## 2. SSL Model and Mapping

- [x] 2.1 Add `VerificationMode types.String` field to `outputSslModel` struct in `models_ssl.go`
- [x] 2.2 Add `VerificationMode *kbapi.KibanaHTTPAPIsOutputSslVerificationMode` field to `outputSSLAPIModel` struct in `models_ssl.go`
- [x] 2.3 Populate `VerificationMode` in `objectValueToSSL()` from the Terraform object value
- [x] 2.4 Populate `VerificationMode` in `toAPI()` when converting `outputSSLAPIModel` to `*kbapi.KibanaHTTPAPIsOutputSsl`, assigning the enum pointer type expected by the generated client
- [x] 2.5 Add `verificationMode *kbapi.KibanaHTTPAPIsOutputSslVerificationMode` parameter to `sslToObjectValue()` and use it to populate the model
- [x] 2.6 Update the null-check guard in `sslToObjectValue()` so it remains null only when all four fields (`certificate`, `certificateAuthorities`, `key`, `verificationMode`) are nil/empty, using the enum-typed `verificationMode` parameter

## 3. Output Type Read Callers

- [x] 3.1 Update `fromAPIKafkaModel()` in `models_kafka.go` to pass `data.Ssl.VerificationMode` directly to `sslToObjectValue()` as `*kbapi.KibanaHTTPAPIsOutputSslVerificationMode`
- [x] 3.2 Update `fromAPIElasticsearchModel()` in `models_elasticsearch.go` similarly
- [x] 3.3 Update `fromAPILogstashModel()` in `models_logstash.go` similarly
- [x] 3.4 Update `fromAPIRemoteElasticsearchModel()` in `models_remote_elasticsearch.go` similarly

## 4. Tests

- [x] 4.1 Update `Test_objectValueToSSL` in `models_ssl_test.go` to cover `verification_mode` round-trip
- [x] 4.2 Update `Test_sslToObjectValue` in `models_ssl_test.go` with cases for `verification_mode` present and absent
- [x] 4.3 Update `Test_objectValueToSSLUpdate` in `models_ssl_test.go` if needed

## 5. Verification

- [x] 5.1 Run `make build` to confirm the project compiles without errors
- [x] 5.2 Run unit tests: `go test ./internal/fleet/output/...`
