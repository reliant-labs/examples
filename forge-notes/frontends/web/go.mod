// Nested module boundary so `go test ./...` from the project root
// skips node_modules and other frontend assets. This frontend has no
// first-party Go code; the module exists solely as a boundary marker.
module example.com/notes/frontends/web

go 1.27.4
