//go:build e2e
// +build e2e

package notes_test

import (
	"fmt"

	"github.com/reliant-labs/forge/pkg/pgtest"

	pb "example.com/notes/gen/services/notes/v1"
)

// freePort finds an available TCP port on loopback for the service binary
// this harness is about to spawn. pgtest.ReserveLoopbackPort is the shared
// implementation — the same one that picks a port for ephemeral postgres.
func freePort() int {
	port, err := pgtest.ReserveLoopbackPort()
	if err != nil {
		panic(fmt.Sprintf("failed to find free port: %v", err))
	}
	return port
}

// Ensure the generated protobuf package import remains referenced.
var _ = pb.File_services_notes_v1_notes_proto
