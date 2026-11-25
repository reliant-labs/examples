// yours: scaffolded once, never touched again — forge will not overwrite this file
//
// Drives the real generated stack (handlers → pkg/crud → internal/db →
// database) against the project's own migrations, asserting real CRUD
// semantics. Extend it as your entities grow. See the forge `testing` skill.
package notes_test

import (
	"context"
	"strings"
	"testing"

	connect "connectrpc.com/connect/v2"
	"github.com/reliant-labs/forge/pkg/orm"
	"github.com/reliant-labs/forge/pkg/svcerr"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	pb "example.com/notes/gen/services/notes/v1"
	notes "example.com/notes/internal/handlers/notes"
	"example.com/notes/pkg/middleware"
)

// crudTestDB returns a fresh database cloned from the template the project's
// own migrations (db/migrations) were applied to once per test binary —
// the schema the generated CRUD code runs against in production.
func crudTestDB(t *testing.T) orm.Context {
	t.Helper()
	return notes.NewMigratedTestDB(t)
}

// crudTestCtx returns a context with authenticated test claims attached —
// what the auth middleware would have produced for a real request.
func crudTestCtx() context.Context {
	return middleware.ContextWithClaims(context.Background(),
		&middleware.Claims{UserID: "test-user", Email: "test@example.com"})
}

// assertCleanError asserts a client-facing error: the expected code, and
// no leaked SQL/driver internals in the message.
//
// The code is svcerr.Code, the code the client is sent — not connect.CodeOf.
// These tests call the handler directly, without the svcerr interceptor, so a
// handler that returns a bare kind (`return nil, svcerr.PermissionDenied(…)`)
// hands back no *connect.Error, and connect.CodeOf would read it as Unknown
// while the generated CRUD paths' *connect.Error read correctly.
func assertCleanError(t *testing.T, label string, err error, want connect.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected %v error, got nil", label, want)
	}
	if got := svcerr.CodeV2(err); got != want {
		t.Errorf("%s: code = %v, want %v (err: %v)", label, got, want, err)
	}
	for _, leak := range []string{"sql:", "SELECT", "no rows in result set", "no such column"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("%s: client-visible error leaks driver/SQL text (%q): %v", label, leak, err)
		}
	}
}

// TestCRUD_Note_Lifecycle exercises the full executed CRUD
// path for Note end to end.
func TestCRUD_Note_Lifecycle(t *testing.T) {
	db := crudTestDB(t)
	svc := notes.NewTestNotes(t, notes.WithDB(db))
	ctx := crudTestCtx()

	// Two creates must produce two rows with distinct, non-empty IDs.
	//
	// The requests come from NewCreateNoteRequest, which forge
	// REGENERATES from the applied schema on every `forge generate`
	// (factories_gen_test.go, beside this file): it seeds the foreign-key
	// parents the request references and fills every field with a value the
	// current constraints accept. A migration you edit later — a GENERATED
	// column, a one-way status CHECK — flows into it instead of breaking this
	// file. Variants 0 and 1 are distinct rows; override any field you assert
	// on (`req.Name = "..."`) after the call.
	first, err := svc.CreateNote(ctx, notes.NewCreateNoteRequest(t, db, 0))
	if err != nil {
		t.Fatalf("create #1: %v", err)
	}
	second, err := svc.CreateNote(ctx, notes.NewCreateNoteRequest(t, db, 1))
	if err != nil {
		t.Fatalf("create #2: %v", err)
	}
	id1 := first.GetNote().GetId()
	id2 := second.GetNote().GetId()
	if id1 == "" || id2 == "" {
		t.Fatalf("create returned empty id(s): %q, %q", id1, id2)
	}
	if id1 == id2 {
		t.Fatalf("two creates returned the same id %q — create must be an insert, never an upsert", id1)
	}
	if first.GetNote().GetCreatedAt() == nil {
		t.Errorf("created_at not set on create (timestamps: true)")
	}

	// Get returns the created row.
	got, err := svc.GetNote(ctx, &pb.GetNoteRequest{Id: id1})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.GetNote().GetId() != id1 {
		t.Errorf("get returned id %q, want %q", got.GetNote().GetId(), id1)
	}

	// List sees both rows.
	listed, err := svc.ListNotes(ctx, &pb.ListNotesRequest{
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if n := len(listed.GetNotes()); n != 2 {
		t.Errorf("list returned %d rows, want 2", n)
	}

	// Update actually mutates: a re-read observes the change.
	row := got.GetNote()
	row.Title = "lifecycle-updated"
	if _, err := svc.UpdateNote(ctx, &pb.UpdateNoteRequest{Note: row}); err != nil {
		t.Fatalf("update: %v", err)
	}
	reread, err := svc.GetNote(ctx, &pb.GetNoteRequest{Id: id1})
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if v := reread.GetNote().GetTitle(); v != "lifecycle-updated" {
		t.Errorf("update did not mutate: Title = %q, want %q", v, "lifecycle-updated")
	}

	// AIP-134 masked update: the mask names ONLY title,
	// so that field changes and nothing else does. The unmasked update
	// above is the documented full-object replace; this is the partial
	// path. See https://google.aip.dev/134.
	maskedRow := reread.GetNote()
	// keptUnmaskedValue uses a fixed, field-independent local name on
	// purpose: deriving it from the field name would produce identifiers
	// ending in an initialism (…Id, …Api) and trip revive's var-naming rule
	// on this owned scaffold.
	keptUnmaskedValue := maskedRow.GetBody()
	// Clobber value the mask does NOT name — it must not persist.
	maskedRow.Body = "MUST-NOT-PERSIST"
	maskedRow.Title = "masked-updated"
	if _, err := svc.UpdateNote(ctx, &pb.UpdateNoteRequest{
		Note:       maskedRow,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title"}},
	}); err != nil {
		t.Fatalf("masked update: %v", err)
	}
	afterMasked, err := svc.GetNote(ctx, &pb.GetNoteRequest{Id: id1})
	if err != nil {
		t.Fatalf("get after masked update: %v", err)
	}
	if v := afterMasked.GetNote().GetTitle(); v != "masked-updated" {
		t.Errorf("masked update did not write the masked field: title = %q, want %q", v, "masked-updated")
	}
	if v := afterMasked.GetNote().GetBody(); v != keptUnmaskedValue {
		t.Errorf("masked update CLOBBERED an unmasked field: body = %q, want %q — update_mask must restrict the write", v, keptUnmaskedValue)
	}

	// An unknown mask path is the caller's mistake: clean InvalidArgument.
	_, err = svc.UpdateNote(ctx, &pb.UpdateNoteRequest{
		Note:       maskedRow,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"this_field_does_not_exist"}},
	})
	assertCleanError(t, "masked update with unknown path", err, connect.CodeInvalidArgument)

	// Update of an id that does not exist is a NotFound, not a 200 echoing
	// back a row the database does not have. Update writes straight through
	// — there is no read before it — so the row count is the only evidence
	// that anything was written.
	ghost := reread.GetNote()
	ghost.Id = "never-existed"
	_, err = svc.UpdateNote(ctx, &pb.UpdateNoteRequest{Note: ghost})
	assertCleanError(t, "update missing", err, connect.CodeNotFound)

	// Delete, then a get of the deleted id maps to a clean NotFound.
	if _, err := svc.DeleteNote(ctx, &pb.DeleteNoteRequest{Id: id1}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = svc.GetNote(ctx, &pb.GetNoteRequest{Id: id1})
	assertCleanError(t, "get after delete", err, connect.CodeNotFound)

	// Get and Delete must AGREE about what "this row does not exist" means.
	// Deleting the same id twice, and deleting an id that never existed,
	// both answer exactly what Get answers for those ids. Delete used to
	// return 200 with an empty body for both, which a client could not
	// tell apart from a delete that did something.
	_, err = svc.DeleteNote(ctx, &pb.DeleteNoteRequest{Id: id1})
	assertCleanError(t, "delete twice", err, connect.CodeNotFound)
	_, err = svc.DeleteNote(ctx, &pb.DeleteNoteRequest{Id: "never-existed"})
	assertCleanError(t, "delete missing", err, connect.CodeNotFound)

	// A never-existed id maps to a clean NotFound too.
	_, err = svc.GetNote(ctx, &pb.GetNoteRequest{Id: "never-existed"})
	assertCleanError(t, "get missing", err, connect.CodeNotFound)
}
