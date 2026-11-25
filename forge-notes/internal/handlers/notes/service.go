package notes

import (
	"fmt"
	"log/slog"
	"net/http"

	connect "connectrpc.com/connect/v2"

	pb "example.com/notes/gen/services/notes/v1"
	notesv1connect "example.com/notes/gen/services/notes/v1/notesv1connect"
	"example.com/notes/pkg/config"

	"github.com/reliant-labs/forge/pkg/orm"
)

// Deps holds dependencies for the notes service. Composition
// (internal/app/compose.go) supplies Logger/Config and resolves added fields
// by type. The rpc methods on *Service are where this service's logic lives;
// see the forge `api` / `services` skill.
type Deps struct {
	DB     orm.Context
	Logger *slog.Logger
	Config *config.Config
	// Add your dependencies here (e.g. Repo Repository). Required ones get
	// a validateDeps() check below. Tag intentionally-optional fields with
	// the `forge:optional-dep` directive directly above the field.
}

// validateDeps returns an error if any required Deps field is nil. The
// gate runs once at construction so handlers need no per-RPC nil-checks.
func (d Deps) validateDeps() error {
	if d.Logger == nil {
		return fmt.Errorf("notes: Deps.Logger is required")
	}
	if d.Config == nil {
		return fmt.Errorf("notes: Deps.Config is required")
	}
	if d.DB == nil {
		return fmt.Errorf("Deps.DB is required by the generated CRUD handlers — set DATABASE_URL (the generated bootstrap constructs the ORM client from it)")
	}
	// Add checks for your required Deps fields here. Example:
	//   if d.Repo == nil { return fmt.Errorf("notes: Deps.Repo is required") }
	return nil
}

// Service implements the notes Connect RPC service.
type Service struct {
	notesv1connect.UnimplementedNotesServiceHandler
	deps Deps
}

// New creates a new notes service instance with the supplied Deps.
// Returns an error when validateDeps rejects a nil required field.
func New(deps Deps) (*Service, error) {
	if err := deps.validateDeps(); err != nil {
		return nil, err
	}
	return &Service{deps: deps}, nil
}

// Name returns the service name for logging and registry.
func (s *Service) Name() string {
	return "notes"
}

// RegisterV2 registers the service's Connect handlers on server
// (mountkit.RegistrarV2). The mount owns the server — its interceptor
// chain and the HTTP mux it is served on — so this only names the rpcs.
func (s *Service) RegisterV2(server *connect.Server) {
	notesv1connect.RegisterNotesServiceHandler(server, s)
}

// RegisterHTTP registers plain HTTP handlers (callbacks, REST endpoints)
// outside Connect RPC. The stack provides recovery/logging/audit; auth is
// per-route. Override to add endpoints — forge.yaml webhooks are wired
// automatically (do not call RegisterWebhookRoutes here). See the forge
// `api` skill.
func (s *Service) RegisterHTTP(_ *http.ServeMux, _ func(http.Handler) http.Handler) {
	// No HTTP endpoints by default. Example:
	//   mux.Handle("POST /api/callbacks/oauth", stack(http.HandlerFunc(s.HandleOAuthCallback)))
}

// Ensure the generated protobuf package import remains referenced even before methods exist.
var _ = pb.File_services_notes_v1_notes_proto
