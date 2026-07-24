# ADR-0001: архитектура FinTask

Статус: принято, 24 июля 2026.

## Контекст

FinTask должен демонстрировать production-подходы на собеседовании, но оставаться
понятным одному разработчику и развертываться одним Docker Compose.

## Решения

### Модульный монолит

Features `identity`, `tasks`, `wallet`, `payments`, `notifications` находятся в
одном процессе и одной PostgreSQL. Границы поддерживаются интерфейсами service и
repository/provider. Микросервисы и Kafka не добавлены: независимого
масштабирования и отдельного владения командами пока нет.

### Cookie и JWT

React использует `Secure`, `HttpOnly`, `SameSite` cookie-сессию. Изменяющие
запросы требуют отдельный CSRF-токен. CLI использует 15-минутный Ed25519 JWT и
одноразовый refresh token. Так browser token не доступен JavaScript, а CLI не
эмулирует cookie jar.

### Double-entry ledger

Баланс кошелька ускоряет чтение, ledger объясняет каждое изменение. PostgreSQL
deferred constraint trigger проверяет минимум две проводки и нулевую сумму при
commit. Суммы представлены `int64` копеек.

### Transactional outbox

Финансовая транзакция создает outbox event, но не вызывает SMTP. Worker доставляет
уведомления после commit с retry и dead-letter. Поэтому сбой внешнего канала не
откатывает деньги.

## Последствия

- Один deployable artifact и простые локальные транзакции.
- Feature можно выделить позже, если появится измеримая причина.
- PostgreSQL является критической зависимостью и требует backup/monitoring.
- Email имеет at-least-once семантику. Ledger credit выполняется не более одного
  раза относительно `provider_payment_id`; после принятого `succeeded` webhook
  результат effectively-once. Reconciliation в первый релиз не входит.
