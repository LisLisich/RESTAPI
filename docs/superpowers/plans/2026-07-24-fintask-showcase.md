# FinTask Showcase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development` or `superpowers:executing-plans`.
> Steps use checkbox syntax for tracking.

**Goal:** Выпустить проверяемый FinTask с identity, платежным ledger,
уведомлениями, React UI, HTTPS, оформленным GitHub и резюме.

**Architecture:** Модульный Go-монолит следует границам
`transport/http -> service -> repository/provider -> domain`. PostgreSQL
остается источником истины; внешние интеграции находятся за интерфейсами.

**Tech Stack:** Go 1.25, net/http, pgx/PostgreSQL, React/TypeScript/Vite,
Docker Compose, Caddy, GitHub Actions.

## Global Constraints

- Новое поведение разрабатывается через red-green-refactor.
- Все новые миграции идут после `000001`; legacy-данные не удаляются.
- Секреты и персональные данные не коммитятся.
- Комментарии в коде пишутся на русском только для неочевидных решений.
- Перед каждым commit выполняются focused tests и `git diff --check`.

---

### Task 1: Identity foundation

**Produces:** account domain, Argon2id password credentials, email tokens,
server sessions, CSRF и `/api/v2` principal.

- [ ] Добавить failing domain/service tests для регистрации и сессии.
- [ ] Добавить account/session migrations и PostgreSQL adapters.
- [ ] Реализовать register, verify, login, logout и password reset.
- [ ] Добавить owner-scoped users/tasks/statistics v2 tests.
- [ ] Запустить `go test ./...`, `go build ./cmd/todoapp`, `go vet ./...`.

### Task 2: Google OIDC and CLI tokens

**Produces:** Google provider port, OIDC callback, Ed25519 access JWT,
rotating refresh token и Go CLI.

- [ ] Добавить fake OIDC contract tests для state, nonce и claims.
- [ ] Реализовать Google start/callback без автоматического linking по email.
- [ ] Добавить token rotation/reuse tests и JWT implementation.
- [ ] Реализовать CLI login/refresh smoke flow.

### Task 3: Wallet and ledger

**Produces:** `Money`, RUB wallet, double-entry ledger и balance/history API.

- [ ] Добавить failing tests для parsing Money и ledger balance.
- [ ] Добавить ledger migrations с deferred zero-sum constraint.
- [ ] Реализовать wallet service/repository/HTTP layers.
- [ ] Проверить несбалансированную транзакцию на реальном PostgreSQL.

### Task 4: YooKassa payments

**Produces:** payment state machine, provider adapter, create/get endpoints и
идемпотентный webhook.

- [ ] Добавить fake-provider tests и конкурентный duplicate-webhook test.
- [ ] Реализовать create payment с `Idempotency-Key`.
- [ ] Реализовать provider status verification и atomic credit.
- [ ] Выполнить YooKassa sandbox smoke test при наличии credentials.

### Task 5: Notifications and outbox

**Produces:** PostgreSQL outbox, in-app notifications, SMTP worker, retry и
dead-letter status.

- [ ] Добавить tests для уникального event effect и retry policy.
- [ ] Реализовать outbox claim через `FOR UPDATE SKIP LOCKED`.
- [ ] Реализовать notification API и SMTP adapter.
- [ ] Проверить доставку через Mailpit.

### Task 6: React and deployment

**Produces:** React UI, Caddy HTTPS deployment, health checks и CI.

- [ ] Создать React/Vite приложение и component tests.
- [ ] Реализовать auth, tasks, wallet, payment и notifications screens.
- [ ] Добавить Playwright happy path.
- [ ] Добавить Docker Compose, Caddy и GitHub Actions.
- [ ] Выполнить VPS deployment после получения домена и SSH-доступа.

### Task 7: GitHub and resume

**Produces:** публичный `fintask`, profile README, release `v1.0.0`, PDF и
hh.ru resume text.

- [ ] Обновить README, ADR, threat model, topics и release notes.
- [ ] Создать profile repository после получения точного имени и контактов.
- [ ] Архивировать старые repositories только после отдельного подтверждения.
- [ ] Создать локальные career artifacts без публикации персональных данных.
- [ ] Проверить каждое утверждение резюме через CI, source или demo.

