//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package connectors -o ./generated/connectors/connectors.gen.go -generate "types,client" ./generated/connectors/bundled.yaml

package main
