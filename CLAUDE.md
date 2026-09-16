# УК-инциденты — мини-апп MAX

Хакатон, демо 2026-09-30. Команда: 1 бек (Go), 1 фронт (React).

- Спека: `docs/superpowers/specs/2026-09-16-uk-incidents-app-design.md`
- Роадмап фич с датами: `docs/superpowers/plans/2026-09-16-features-roadmap.md`
- Планы: П1 `plan-1-backend-core-incidents.md`, П2 `plan-2-backend-admin-events.md`, П3 `plan-3-frontend.md` (там же)

## Раскладка репо

- `backend/` — Go-модуль `ukapp`: `core/` (чистая логика, без БД/HTTP), `api/` (реализация HTTP-хендлеров, БД, MAX-клиент, outbox-воркер), `gen/` (сгенерированное, руками не править), `cmd/api/main.go`. Пути вида `core/...`, `api/...` в планах — относительно `backend/`.
- `frontend/` — Vite + React + TS: `src/features/{onboarding,incidents,events,uk-panel}`, `src/shared/{api,max}`.

## Кодогенерация (источники правды)

- **HTTP-контракт — `backend/openapi.yaml`.** `make gen-api` → `gen/api` (oapi-codegen, strict-server на stdlib `net/http`). Пакет `api` реализует `oapi.StrictServerInterface`; роуты/JSON-типы руками не пишем. Меняешь ручку → меняешь `openapi.yaml` в том же коммите; фронт генерирует TS-типы из того же файла.
- **Схема БД — `backend/api/schema.sql`.** `make gen-db` → применяет схему в локальный Postgres (`docker compose`) и генерирует `gen/db` (go-jet: `model/` структуры + `table/` билдер). Запросы пишем через jet, не строками. Схема применяется при старте приложения тем же `schema.sql`.
- `make gen` = оба. Инструменты закреплены в `go.mod` (`tool` directive), ставить ничего не надо.
- Планы П1/П2 писались до кодогена: их `http.ServeMux`-роуты, ручные JSON-структуры и строковый SQL через `pgxpool` при выполнении заменяем на `gen/api` + jet. Логика хендлеров, тесты и SQL-семантика — из планов.

## Ограничения (constitution)

- Go ≥ 1.26. Внешние зависимости: `pgx/v5` (драйвер через `pgx/v5/stdlib` + `database/sql`), `go-jet/jet/v2`, `oapi-codegen/runtime`. Всё остальное — stdlib. Новую зависимость — только с обоснованием в PR.
- Ошибки API — JSON `{"error": "..."}` + HTTP-код (не `http.Error` text/plain из планов). Логи — `log/slog`.
- Бизнес-событие и запись в `outbox_message` — в одной транзакции.
- MAX Bot API: `https://platform-api2.max.ru`, заголовок `Authorization: <BOT_TOKEN>`, ≤ 30 rps.
- Валидация initData: `secret = HMAC_SHA256(key="WebAppData", data=BOT_TOKEN)`; строка — пары `key=value` без `hash`, значения URL-декодированы, ключи отсортированы, разделитель `\n`; `hash == hex(HMAC_SHA256(key=secret, data=строка))`. Без валидной подписи — 401.
- Дедуп: тот же `house_id` + та же категория + открытый (`new`/`in_progress`) + не старше 2 ч.
- Порог спора: `no_votes / subscribers >= 0.30` (от всех подписчиков, не от ответивших); голоса `shadow_banned` не считаются. Окно ответа 48 ч. Теневой бан после 3 ложных отклонений.
- Категории: `electricity | water | elevator | yard | other`. Роли: `resident | uk_dispatcher`.
- Env: `DATABASE_URL`, `MAX_BOT_TOKEN`, `DADATA_TOKEN`, `PORT` (default 8080).
- Линт: `cd backend && make lint` (golangci-lint v2, `~/go/bin`). Архитектурные границы `cmd → api → core` и белый список зависимостей проверяет depguard в `backend/.golangci.yml` — перед коммитом линт должен быть зелёным.
- Тесты: `core` — юнит без моков; `api` — против реального Postgres через `DATABASE_URL` (skip, если не задан); фронт — ручной прогон в MAX-песочнице.

## Процесс

- **`frontend/` в этой сессии не трогаем** — его ведёт другой участник по плану П3. Мы делаем только `backend/` (П1, П2).

- Ветка на фичу (`feat/<n>-<name>`), PR в `main`.
- Планы выполняются задача за задачей (`superpowers:executing-plans`), коммит после каждой задачи.
