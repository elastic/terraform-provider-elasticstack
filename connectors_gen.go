//go:generate go run github.com/ogen-go/ogen/cmd/ogen --infer-types --no-server --target generated/connectors -package connectors --clean --debug.ignoreNotImplemented "discriminator inference" ./generated/connectors/bundled.yaml

// go:generate go run github.com/ogen-go/ogen/cmd/ogen --help

package main
