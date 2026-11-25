# Database

The schema is declared here, in SQL, and nowhere else. `forge generate` applies
the migrations to a throwaway postgres and projects what it finds into Go: entity
structs, the `<Entity>Store` ORM and the CRUD wiring under `internal/db/`.

## Layout

```
db/
  migrations/   # <UTC timestamp>_<name>.up.sql — roll forward only
  seeds/        # dev seed vocabulary and hand-authored demo data
```

New migration: `forge db migration new <name>`, then `forge generate`.
Dev data: `forge db seed apply`.

## Queries

The generated store covers reads by key, filtered lists and writes. A query it
cannot express goes in `internal/db/<entity>_repo_ext.go`, which forge scaffolds
beside each entity and never rewrites: compose `orm.QueryOption` filters on the
generated delegates, or drop to SQL with `db.Bun()` — bun takes CTEs,
`FOR UPDATE SKIP LOCKED` and raw statements directly. Cover every such query
with a test against a real migrated database (`testkit`), which is what checks
the SQL against the schema.
