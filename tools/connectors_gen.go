//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -package connectors -o ../generated/connectors/connectors.gen.go -generate "types,client" ../generated/connectors/bundled.yaml
//go:generate go run go.uber.org/mock/mockgen -destination=../generated/alerting/api_alerting_mocks.go -package=alerting -source ../generated/alerting/api_alerting.go AlertingAPI
//go:generate go run go.uber.org/mock/mockgen -destination=../internal/clients/kibana/alerting_mocks.go -package=kibana -source ../internal/clients/kibana/alerting.go ApiClient

package tools
