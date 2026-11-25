# db/seeds/custom — your demo-data overlay

Hand-authored (or agent-authored) demo data. Applied AFTER the synthesized
dataset, in lexicographic order, by `forge db seed apply`, `forge db reset` and the
first-boot seed of `forge env up`. This directory is
scaffolded once and owned by you — forge never regenerates it.

Rules:

- Make every row idempotent: a stable key + `ON CONFLICT ... DO NOTHING`.
- Overriding a generated row is sanctioned: the runtime PKs are stable
  deterministic UUIDv5 values, so `ON CONFLICT (id) DO UPDATE SET ...`
  against one lets you re-title/flavor it without forking the dataset.
- A row that violates a constraint (dropped column, bad enum) fails loudly
  at apply — the fix is yours, exactly like a migration.

See example.sql for the shape.
