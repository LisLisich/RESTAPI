# Threat model FinTask

## Активы

- парольные credentials, cookie-сессии, JWT/refresh tokens;
- Google OIDC identity;
- платежный статус, баланс и ledger;
- email и персональные данные пользователя.

## Границы доверия

Browser/CLI, публичный Caddy, Go API, PostgreSQL, Google, YooKassa и SMTP — разные
границы. Webhook и browser input считаются недоверенными.

## Основные угрозы и меры

| Угроза | Мера |
|---|---|
| Кража пароля | Argon2id; пароль не логируется |
| Кража browser token через JS | `HttpOnly`, `Secure`, `SameSite` cookie |
| CSRF | state-changing запросы требуют отдельный CSRF header |
| JWT forgery | Ed25519 signature, `iss`, `aud`, `exp`, `typ`, `sub`, `jti` |
| Refresh replay | одноразовая DB-ротация; использованный token отклоняется |
| OIDC login CSRF/replay | одноразовые `state`, `nonce`, PKCE S256, state cookie |
| Подмена Google identity | JWKS signature, `iss`, `aud`, `exp`, `sub` |
| Account takeover по email | Google email не связывает существующий аккаунт |
| Поддельный webhook | актуальный payment повторно читается через YooKassa API |
| Двойное зачисление | unique provider id, row lock, status guard, DB-транзакция |
| Несбалансированный ledger | domain validation и deferred DB trigger |
| SQL injection | параметры `$1...`, пользовательские значения не конкатенируются |
| Утечка БД | PostgreSQL не имеет публичного порта |
| Утечка callback/токенов через логи | application log пишет только URL path; Caddy редактирует `code` и `state`; authorization/cookie headers скрыты |
| Раскрытие внутренних ошибок | подробная ошибка остается в server log, клиент получает только HTTP status text и короткое message |
| Потеря уведомления | transactional outbox, retry, dead-letter |

## Остаточные риски

- SMTP at-least-once может повторить email после сетевого сбоя.
- Нет MFA, rate limiting и централизованного secret manager.
- Caddy и PostgreSQL требуют регулярных обновлений и проверки backup restore.
- Перед реальными деньгами нужны аудит, мониторинг, reconciliation и возвраты.
- Администраторы логов все равно имеют доступ к техническим метаданным; доступ и
  срок хранения логов должны быть ограничены.
