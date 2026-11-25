# db/seeds

Development seed data is **not** a file in this project. It materializes at
runtime, deterministically, from the applied schema:

    forge db seed apply     # introspect db/migrations, INSERT ~20 FK-coherent
                            # rows per table (idempotent, ON CONFLICT DO NOTHING)
    forge db seed status    # per-table seeded-row counts
    forge db seed reset     # wipe seeded tables and re-seed (dev only)

`forge env up` auto-seeds a fresh dev database on first boot (all tables
empty). apply/reset refuse any non-dev environment — seeding never runs
against staging or production, and the applier is not compiled into your
server binary.

## Domain vocabulary — make seeded rows speak your domain

**`db/seeds/vocab.yaml` is where your domain's words live, and nothing
else supplies them.** Without it the seeder invents values that announce
themselves — `sample_<column>_<n>` — because what a column MEANS is not
in the schema and forge will not guess it from the column's name. A database
full of `sample_*` is forge telling you nobody has filled this in yet.

The file (yours, scaffolded once) maps `table.column` keys to value
pools; matched columns draw from them instead — same determinism, validated
against your schema's constraints. Write it right after authoring migrations,
while the domain is fresh. See the file's header for the schema.

What forge DOES derive on its own, because you declared it: enum and CHECK
vocabularies, `char_length` / varchar caps, numeric ranges, regex
`CHECK (col ~ '...')` patterns, NOT NULL, UNIQUE, and every foreign key.

## Custom overlay — your hand-authored demo data

Put domain-flavored or agent-written demo rows in `db/seeds/custom/`.
Files there apply **after** the runtime dataset, in lexicographic order.
This directory is yours: forge scaffolds it once and never regenerates it.
