-- Migration: add_note_pinned
-- Created: 2026-11-25T23:17:52Z
--
-- The schema is the 2 migration(s) in this directory, applied in order (newest: 20261125231524_create_notes.up.sql). To read it:
--   forge project shapes --kind table          every table, with the file:line declaring it
--   forge db introspect --dsn "$DATABASE_URL"  the schema a database actually has
--   forge db migration new <name> --context    a new file with the schema written into it
--
-- Write your migration SQL below:

-- forge:backward-compatible — adds a column the running release never reads.
ALTER TABLE notes ADD COLUMN pinned BOOLEAN NOT NULL DEFAULT false;
