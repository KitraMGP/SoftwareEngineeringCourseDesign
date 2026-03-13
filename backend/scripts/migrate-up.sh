#!/usr/bin/env sh
set -eu

if [ -f ".env" ]; then
  set -a
  . ./.env
  set +a
fi

if [ -f ".env.local" ]; then
  set -a
  . ./.env.local
  set +a
fi

if [ -z "${DATABASE_DSN:-}" ]; then
  echo "DATABASE_DSN is required. Create backend/.env first." >&2
  exit 1
fi

for file in migrations/*.sql; do
  echo "Applying migration: $file"
  awk '
    /^-- \+goose Up/ { in_up=1; next }
    /^-- \+goose Down/ { in_up=0 }
    in_up { print }
  ' "$file" | psql "$DATABASE_DSN" -v ON_ERROR_STOP=1 -f -
done

echo "All migrations applied."
