#!/bin/sh
set -e
psql -h "$PGHOST" -U "$PGUSER" -d "$PGDATABASE" -c "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY)"
for f in /migrations/*.sql; do
  name=$(basename "$f")
  if [ "$(psql -h "$PGHOST" -U "$PGUSER" -d "$PGDATABASE" -tAc "SELECT 1 FROM schema_migrations WHERE filename='$name'")" = "1" ]; then
    echo "skip $name"
    continue
  fi
  echo "apply $name"
  psql -h "$PGHOST" -U "$PGUSER" -d "$PGDATABASE" -v ON_ERROR_STOP=1 -f "$f"
  psql -h "$PGHOST" -U "$PGUSER" -d "$PGDATABASE" -c "INSERT INTO schema_migrations (filename) VALUES ('$name')"
done
