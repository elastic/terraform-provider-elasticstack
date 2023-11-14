package tools

//go:generate go run ../generated/fleet/getschema.go -v v8.10.0 -o ../generated/fleet/fleet-filtered.json
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen -package=fleet -generate=types,client -o ../generated/fleet/fleet.gen.go ../generated/fleet/fleet-filtered.json
