# FinTask

[![CI](https://github.com/LisLisich/RESTAPI/actions/workflows/ci.yml/badge.svg?branch=fintask-showcase)](https://github.com/LisLisich/RESTAPI/actions/workflows/ci.yml)

FinTask — учебный финансовый backend на Go: задачи, регистрация, Google OIDC,
RUB-кошелек, sandbox-платежи YooKassa и уведомления. Проект построен как
модульный монолит с PostgreSQL-транзакциями и React-интерфейсом.

![Рабочее пространство FinTask](docs/assets/fintask-workspace.png)

## Что реализовано

- регистрация, Argon2id, подтверждение email и сброс пароля;
- защищенные cookie-сессии с CSRF для браузера;
- Ed25519 access JWT и ротация refresh-токенов для `fintaskctl`;
- Google Authorization Code + PKCE/OIDC с проверкой подписи, `state`, `nonce`,
  `iss`, `aud`, `exp` и `sub`;
- owner-scoped CRUD задач в `/api/v2`;
- `Money` как `int64` копеек, RUB-кошелек и double-entry ledger;
- YooKassa sandbox с клиентской и provider-идемпотентностью;
- PostgreSQL outbox, in-app и SMTP-уведомления, retry и dead-letter;
- React + TypeScript, Vitest и Playwright для desktop/mobile;
- Caddy HTTPS, healthcheck, backup, rollback и GitHub Actions.

## Архитектура

```mermaid
flowchart LR
    UI["React / browser"] -->|cookie + CSRF| HTTP["Go HTTP transport"]
    CLI["fintaskctl"] -->|Ed25519 JWT| HTTP
    HTTP --> ID["identity"]
    HTTP --> TASKS["tasks"]
    HTTP --> WALLET["wallet"]
    HTTP --> PAY["payments"]
    PAY --> YK["YooKassa sandbox"]
    ID --> DB[("PostgreSQL")]
    TASKS --> DB
    WALLET --> DB
    PAY -->|"payment + wallet + ledger + outbox\nодна транзакция"| DB
    DB --> WORKER["outbox worker"]
    WORKER --> SMTP["SMTP"]
    WORKER --> NOTIFY["in-app notifications"]
```

Направление зависимостей внутри feature:

```text
transport/http -> service -> repository/provider interfaces -> adapters
```

Ключевые решения описаны в [ADR](docs/architecture/0001-fintask-decisions.md),
[threat model](docs/security/threat-model.md) и
[payment sequence](docs/architecture/payment-sequence.md).

## Финансовые инварианты

- Денежные суммы не используют `float`.
- Валюта релиза — только `RUB`.
- Ledger-транзакция содержит минимум две проводки с нулевой суммой.
- Один `provider_payment_id` зачисляется не более одного раза.
- Webhook не считается доверенным: статус повторно запрашивается у YooKassa.
- Платеж, кошелек, ledger и outbox меняются одной DB-транзакцией.
- Ошибка SMTP не откатывает подтвержденное зачисление.

## Быстрый запуск

Требования: Go 1.25, Node.js 22, Docker Compose.

```bash
cp .env.example .env
go run ./cmd/fintaskctl keygen
```

Заполните в `.env` `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` и
`JWT_ED25519_PRIVATE_KEY` полученной строкой. Для локального запуска оставьте
YooKassa, Google и SMTP выключенными.

```bash
set -a
. ./.env
set +a
docker compose up -d todoapp-postgres
docker compose run --rm todoapp-postgres-migrate \
  -path /migrations \
  -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@todoapp-postgres:5432/${POSTGRES_DB}?sslmode=disable" up
docker compose --profile dev up -d --build todoapp caddy api-forwarder
```

Откройте `https://localhost`. Для публичного VPS задайте домен в `DOMAIN`;
Caddy автоматически получает и обновляет TLS-сертификат. Локальный сертификат
Caddy может потребовать ручного доверия. Профиль `dev` дополнительно публикует
API только на `127.0.0.1:5050`, поэтому команды CLI ниже работают с адресом по
умолчанию. На VPS профиль `dev` не включается.

Подробности: [deployment runbook](docs/operations/deployment.md).

## CLI

Секреты читаются интерактивно без отображения и удаляются из окружения после
команды. Не передавайте пароль или токен прямо в командной строке.

```bash
read -rsp "FinTask password: " FINTASK_PASSWORD
echo
export FINTASK_PASSWORD
go run ./cmd/fintaskctl login --email user@example.com
unset FINTASK_PASSWORD

read -rsp "FinTask access token: " FINTASK_ACCESS_TOKEN
echo
export FINTASK_ACCESS_TOKEN
go run ./cmd/fintaskctl wallet
unset FINTASK_ACCESS_TOKEN

read -rsp "FinTask refresh token: " FINTASK_REFRESH_TOKEN
echo
export FINTASK_REFRESH_TOKEN
go run ./cmd/fintaskctl refresh
unset FINTASK_REFRESH_TOKEN
```

Для публичного стенда задайте `FINTASK_API_URL=https://<domain>` через
окружение или используйте флаг `--api-url`.

## Проверки

```bash
go test -race ./...
go vet ./...
govulncheck ./...

cd web
npm ci
npm audit --audit-level=moderate
npm test
npm run build
npm run test:e2e
```

На текущем срезе: **269 тестовых функций**, **86 тестовых файлов** и **62,3%**
общего statement coverage. CI дополнительно проверяет миграции PostgreSQL
`up -> down -> up` и собирает Docker image.

## Безопасность

- пароли хешируются Argon2id;
- session/refresh/state tokens хранятся только как хеши;
- browser mutations требуют CSRF-токен;
- JWT подписываются Ed25519 и живут 15 минут;
- Google identity определяется по неизменяемому `sub`, а не по email;
- совпавший Google email не связывает аккаунты автоматически;
- PostgreSQL не публикуется наружу;
- секреты не входят в Git и передаются через environment.

## Ограничения релиза

- YooKassa работает только в тестовом магазине.
- Нет возвратов, переводов, вывода, MFA и административных ролей.
- SMTP имеет at-least-once семантику: при редком сбое подтверждения возможен
  повтор email, но не повтор финансовой проводки.
- `/api/v1` сохранен для legacy users/tasks; новые сценарии используют
  owner-scoped `/api/v2`.
- Kafka и микросервисы намеренно исключены: текущая нагрузка не оправдывает
  операционную сложность.

## Документация

- Swagger UI: `/swagger/`;
- [развертывание и эксплуатация](docs/operations/deployment.md);
- [архитектурные решения](docs/architecture/0001-fintask-decisions.md);
- [последовательность платежа](docs/architecture/payment-sequence.md);
- [модель угроз](docs/security/threat-model.md).

Лицензия: [MIT](LICENSE).
