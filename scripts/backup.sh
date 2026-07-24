#!/usr/bin/env sh
set -eu

backup_dir="${BACKUP_DIR:-./backups}"
retention_days="${BACKUP_RETENTION_DAYS:-7}"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"

mkdir -p "$backup_dir"
docker compose exec -T todoapp-postgres \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom \
  > "$backup_dir/fintask-$timestamp.dump"

find "$backup_dir" -type f -name 'fintask-*.dump' -mtime "+$retention_days" -delete
printf 'Backup created: %s\n' "$backup_dir/fintask-$timestamp.dump"
