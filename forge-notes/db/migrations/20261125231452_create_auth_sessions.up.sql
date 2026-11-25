-- forge:backward-compatible — a new table no older release reads.
-- Server-side sign-in sessions (forge/pkg/session.SchemaSQL): the browser's cookie
-- holds an opaque id; each row holds that session's tokens, sealed under a key
-- derived from the id, so a copy of this table signs no one in.
-- This migration is YOURS from birth: forge writes it once and never rewrites
-- it. Evolve the table with a NEW migration.

CREATE TABLE auth_sessions (
    id_hash         bytea       PRIMARY KEY,
    subject         text        NOT NULL,
    sealed          bytea       NOT NULL,
    user_agent      text        NOT NULL DEFAULT '',
    issuer_session  text        NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    proved_at       timestamptz NOT NULL DEFAULT now(),
    last_seen_at    timestamptz NOT NULL DEFAULT now(),
    refreshed_at    timestamptz NOT NULL DEFAULT now(),
    idle_expires_at timestamptz NOT NULL,
    expires_at      timestamptz NOT NULL
);

CREATE INDEX auth_sessions_subject_idx ON auth_sessions (subject, created_at DESC);
CREATE INDEX auth_sessions_idle_expires_idx ON auth_sessions (idle_expires_at);
CREATE INDEX auth_sessions_expires_idx ON auth_sessions (expires_at);

COMMENT ON COLUMN auth_sessions.id_hash IS 'forge:secret';
COMMENT ON COLUMN auth_sessions.sealed IS 'forge:secret';

-- Sign-in handoffs (forge/pkg/session.HandoffSchemaSQL): the one-time requests and codes
-- a signed-in session hands to a desktop app or sibling host, sealed the same way.
-- Unused (and empty) until handoff_redirects turns the handoff on.

CREATE TABLE auth_handoffs (
    key_hash   bytea       PRIMARY KEY,
    sealed     bytea       NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX auth_handoffs_expires_idx ON auth_handoffs (expires_at);

COMMENT ON COLUMN auth_handoffs.key_hash IS 'forge:secret';
COMMENT ON COLUMN auth_handoffs.sealed IS 'forge:secret';
