-- Born from proto message services.notes.v1.Note (`forge scaffold entity --from-proto`).
-- This migration is YOURS from birth: forge never re-reads the proto to
-- regenerate or edit it. Evolution is a NEW migration plus a proto edit —
-- the schema truth and the wire truth evolve on independent clocks.
CREATE TABLE notes (
    id TEXT PRIMARY KEY CHECK (id <> ''),
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);

-- Ownership: not declared. Say whose rows these are (an owner declaration on
-- the table, which scopes every generated CRUD op to the caller) or why no caller
-- owns them, in a new migration; until then the crud-table-unowned lint flags this
-- table (see the forge `db/crud-overrides` skill).
