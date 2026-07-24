# Deployment runbook

## VPS

Ubuntu VPS должен публиковать только SSH `22/tcp`, HTTP `80/tcp` и HTTPS
`443/tcp+udp`. PostgreSQL и Go API доступны только внутри Compose network.

1. Создать DNS `A` и при использовании IPv6 `AAAA`, указывающие `DOMAIN` на VPS.
2. Открыть firewall только для `22/tcp`, `80/tcp`, `443/tcp` и `443/udp`;
   входящие `80/443` нужны Caddy для ACME и HTTPS.
3. Установить Docker Engine и Compose plugin.
4. Клонировать repository и создать `.env` из `.env.example`.
5. Задать `DOMAIN`, PostgreSQL credentials и Ed25519 key.
6. Включать Google/YooKassa/SMTP только после добавления соответствующих secret.
7. Запустить миграции, затем `docker compose up -d todoapp caddy`.
8. Проверить DNS, сертификат и `https://<domain>/healthz` извне VPS.

```bash
set -eu
set -a
. ./.env
set +a
docker compose up -d todoapp-postgres
docker compose run --rm todoapp-postgres-migrate \
  -path /migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todoapp-postgres:5432/${POSTGRES_DB}?sslmode=disable" up
docker compose up -d todoapp caddy
docker compose ps
```

## Backup

```bash
set -a
. ./.env
set +a
BACKUP_DIR=/srv/fintask/backups ./scripts/backup.sh
```

Backup считается рабочим только после периодического тестового restore в
отдельную PostgreSQL.

### Проверка восстановления

Команды создают отдельную БД и не меняют рабочую. Имя dump передается явно.

```bash
set -eu
set -a
. ./.env
set +a
RESTORE_DB="${POSTGRES_DB}_restore"
DUMP_FILE=/srv/fintask/backups/fintask-YYYYMMDDTHHMMSSZ.dump

docker compose exec -T todoapp-postgres \
  createdb -U "$POSTGRES_USER" "$RESTORE_DB"
docker compose exec -T todoapp-postgres \
  pg_restore -U "$POSTGRES_USER" -d "$RESTORE_DB" --exit-on-error \
  < "$DUMP_FILE"
docker compose exec -T todoapp-postgres \
  psql -U "$POSTGRES_USER" -d "$RESTORE_DB" -c '\dt todoapp.*'
test "$(docker compose exec -T todoapp-postgres \
  psql -U "$POSTGRES_USER" -d "$RESTORE_DB" -Atc \
  "SELECT to_regclass('todoapp.accounts')")" = "todoapp.accounts"
docker compose exec -T todoapp-postgres \
  dropdb -U "$POSTGRES_USER" "$RESTORE_DB"
```

Если проверка завершилась ошибкой, restore-БД не удаляется автоматически:
сначала сохраните вывод и выясните причину.

## Rollback

Images тегируются release-тегами. Откат приложения не откатывает схему
автоматически.

```bash
./scripts/rollback.sh v0.9.0
curl --fail https://<domain>/healthz
```

Down-миграцию выполнять отдельно только после проверки совместимости и backup.

## Диагностика

```bash
docker compose ps
docker compose logs --tail=200 todoapp
docker compose logs --tail=200 caddy
docker compose exec todoapp-postgres pg_isready
```

Секреты, DSN и payload с токенами не добавляются в issue или публичные логи.
