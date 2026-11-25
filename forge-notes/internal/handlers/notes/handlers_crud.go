// yours: scaffolded once, never touched again — forge will not overwrite this file
// Each method delegates to github.com/reliant-labs/forge/pkg/crud (error mapping, cursor
// pagination, page-size clamping, ordering, total counts, update masks); the per-entity wiring is
// regenerated in handlers_crud_ops_gen.go.
// pkg/crud never reads the caller on its own. The interceptor already turned away anyone without a
// valid token (unless the rpc declares auth_required: false; pkg/middleware/procedures_gen.go).
// For a table whose migration declares `forge:owner=`, the generated op asks your
// crud.OwnerResolver (Deps.Owners) who the caller is and scopes every read and write to their
// rows. For every other table, WHO MAY DO WHAT and WHICH ROWS THEY MAY SEE are yours: import
// "example.com/notes/pkg/middleware", resolve the caller with middleware.GetUser(ctx), and wrap
// the op's Filters in a WHERE on those claims. See `forge skill load auth`.
// EACH METHOD'S `Auth:` LINE says which case it is, read from its rpc's auth_required and its
// table's forge:owner declaration:
// AUTHENTICATED, UNSCOPED — only a valid token reaches it, and nothing in its body reads who that
// is yet: the paragraph above is the to-do.
// Customize an RPC by replacing its delegation here; CRUD RPCs added later are appended by `forge
// generate`, and your content is never modified.
package notes

import (
	"context"

	"github.com/reliant-labs/forge/pkg/crud"

	pb "example.com/notes/gen/services/notes/v1"
	"example.com/notes/internal/metrics"
)

// CreateNote delegates CRUD create to pkg/crud.
// Auth: AUTHENTICATED, UNSCOPED — nothing reads the caller; scope with middleware.GetUser(ctx) (see file header).
func (s *Service) CreateNote(
	ctx context.Context,
	req *pb.CreateNoteRequest,
) (*pb.CreateNoteResponse, error) {
	resp, err := crud.Create(s.crudCreateNoteOp())(ctx, req)
	metrics.RecordCreated(ctx, "Note", err)
	return resp, err
}

// GetNote delegates CRUD get to pkg/crud.
// Auth: AUTHENTICATED, UNSCOPED — nothing reads the caller; scope with middleware.GetUser(ctx) (see file header).
func (s *Service) GetNote(
	ctx context.Context,
	req *pb.GetNoteRequest,
) (*pb.GetNoteResponse, error) {
	return crud.Get(s.crudGetNoteOp())(ctx, req)
}

// UpdateNote delegates CRUD update to pkg/crud.
// Auth: AUTHENTICATED, UNSCOPED — nothing reads the caller; scope with middleware.GetUser(ctx) (see file header).
// AIP-134: req.UpdateMask is honored (empty or "*" = full replace); op.ReadOnly columns are refused in it and kept on a full replace.
func (s *Service) UpdateNote(
	ctx context.Context,
	req *pb.UpdateNoteRequest,
) (*pb.UpdateNoteResponse, error) {
	return crud.Update(s.crudUpdateNoteOp())(ctx, req)
}

// DeleteNote delegates CRUD delete to pkg/crud.
// Auth: AUTHENTICATED, UNSCOPED — nothing reads the caller; scope with middleware.GetUser(ctx) (see file header).
func (s *Service) DeleteNote(
	ctx context.Context,
	req *pb.DeleteNoteRequest,
) (*pb.DeleteNoteResponse, error) {
	return crud.Delete(s.crudDeleteNoteOp())(ctx, req)
}

// ListNotes delegates CRUD list to pkg/crud.
// Auth: AUTHENTICATED, UNSCOPED — nothing reads the caller; scope with middleware.GetUser(ctx) (see file header).
func (s *Service) ListNotes(
	ctx context.Context,
	req *pb.ListNotesRequest,
) (*pb.ListNotesResponse, error) {
	return crud.List(s.crudListNotesOp())(ctx, req)
}
