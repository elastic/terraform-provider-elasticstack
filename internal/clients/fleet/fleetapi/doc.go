package fleetapi

//go:generate go run generate.go -v v8.7.1 -o fleet-filtered.json
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package=fleetapi -generate=types -o ./fleetapi_gen.go fleet-filtered.json
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package=fleetapi -generate=client -o ./client_gen.go fleet-filtered.json
