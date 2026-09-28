# УК-инциденты — мини-приложение MAX для заявок в управляющую компанию

Жители дома сообщают об авариях (нет воды, свет, лифт) прямо в мессенджере MAX. Одинаковые
заявки от соседей автоматически склеиваются в одну, УК получает сгруппированный поток, а жители —
пуш-уведомления о каждом изменении статуса.

Прод: https://nerionapp.ru · бот: `t117_hakaton_max_bot`

## Возможности

- Привязка к дому по адресу (подсказки DaData, геолокация) или по QR-наклейке в подъезде.
- Реальная управляющая компания дома подтягивается из ГИС ЖКХ.
- Создание заявки с фото; категория определяется ML-классификатором по тексту.
- Автосклейка дубликатов: ML-модель сравнивает новую заявку с открытыми заявками дома.
  Заявка соседа о той же аварии присоединяется к существующей (ответ `200` вместо `201`),
  диспетчер видит одну карточку со счётчиком жителей.
- Ручное объединение заявок диспетчером.
- Жители подтверждают выполнение — заявка закрывается, когда подтвердило достаточно людей.
- Массовые оповещения дома от УК, пуши о смене статуса со ссылкой на заявку.
- Шеринг заявки в чаты MAX и QR-наклейки для подъездов — жители соседних домов подключаются сами.

## Архитектура

```
MAX (мини-приложение, frontend/)
        │  Authorization: tma <initData>
        ▼
api (Go, backend/) ──── PostgreSQL
  │   │   │   └── laya   (Node.js, классификатор категорий, backend/laya/)
  │   │   └────── dedup  (Python, ML-дедупликация, ml/dedup/)
  │   └────────── mock-uk (Go, mock/uk/) ── PostgreSQL
  │                 └── ЛК диспетчера /lk (канбан-доска)
  └── внешние API: MAX Bot API, DaData, ГИС ЖКХ
```

| Сервис | Каталог | Стек |
|---|---|---|
| `api` | `backend/` | Go 1.26, gin, pgx, go-jet, OpenAPI (`backend/openapi.yaml`) |
| `laya` | `backend/laya/` | Node.js 20, ONNX-модель `convaiinnovations/laya` |
| `dedup` | `ml/dedup/` | Python 3.10, onnxruntime, cross-encoder на базе `sergeyzh/BERTA` |
| `mock-uk` | `mock/uk/` | Go 1.26 — имитация системы УК и веб-кабинет диспетчера |
| `frontend` | `frontend/` | React 19, Vite, `@maxhub/max-ui` |

Список всех зависимостей с лицензиями — [`requirements.txt`](requirements.txt).
Все используемые библиотеки и модели распространяются под open-source лицензиями
(MIT, Apache-2.0, BSD-3-Clause, PostgreSQL License).

## Техническое ограничение MVP

Реальные базы Госуслуг (ЕСИА) и внутренние системы УК закрыты, поэтому для демонстрации мы
развернули собственный mock-сервер УК и кабинет диспетчера (канбан-доску), куда падают
сгруппированные заявки. Обмен с ним идёт по тому же REST-контракту, что потребуется от реальной
УК (`docs/uk-integration.md`) — для подключения настоящей системы меняется только адаптер
`backend/internal/ukclient`.

## Авторизация

| Кто | Как | Где в коде |
|---|---|---|
| Житель | Заголовок `Authorization: tma <initData>`. MAX подписывает `initData` ключом бота, бэкенд проверяет HMAC-SHA256 подпись и срок. Пользователь создаётся при первом запросе. | `backend/internal/domain` (`ValidateInitData`), `backend/internal/transport/http/middleware.go` |
| Диспетчер УК | Вход по ИНН организации и паролю (`POST /api/auth/esia-mock` — имитация входа через ЕСИА). Пароль — bcrypt, 5 неудачных попыток за 15 минут — блокировка. Выдаётся JWT (ES256) с ролью `uk_dispatcher` и id УК, дальше `Authorization: Bearer <jwt>`. | `backend/internal/ukauth`, `backend/internal/service/ukadmin.go` |
| Система УК ↔ api | Статический bearer-токен `UK_API_TOKEN` в обе стороны. | `backend/internal/ukclient`, `mock/uk/internal/api` |

Роли проверяются в middleware: ручки `/api/uk/*` и смена статуса — только диспетчер,
создание/присоединение/подтверждение заявок — только житель; диспетчер видит только дома своей УК.

## Внешние API

| API | Для чего | Ключ |
|---|---|---|
| MAX Bot API (`platform-api2.max.ru`) | пуш-уведомления жителям, проверка `initData` | `MAX_BOT_TOKEN` |
| DaData (`suggestions.dadata.ru`) | подсказки адреса, геокодирование, ФИАС-id дома | `DADATA_TOKEN` |
| ГИС ЖКХ (`dom.gosuslugi.ru`, публичный поиск по адресу) | реальная УК дома, проверка что дом многоквартирный | не нужен; отключается `GISGKH_DISABLED=1` |
| Система УК (в MVP — `mock-uk`) | регистрация заявок, синхронизация статусов | `UK_API_TOKEN` |

## Запуск

Нужны Docker и Docker Compose.

```bash
cp .env.example .env    # заполнить переменные, см. ниже
docker compose pull     # готовые образы из GHCR
docker compose up -d
```

Сборка из исходников вместо GHCR:

```bash
gh release download dedup-ftM2-berta-s16 -p bundle.tar.gz -D ml/dedup   # веса модели дедупликации
docker compose up -d --build
```

После старта:

- API — http://localhost:8080 (`GET /health`), спецификация — `backend/openapi.yaml`;
- кабинет диспетчера УК — http://localhost:8081/lk/login;
- фронтенд: `cd frontend && npm ci && npm run dev` (в проде собирается в CI и отдаётся nginx).

### Переменные окружения (`.env`)

| Переменная | Обязательна | Описание |
|---|---|---|
| `POSTGRES_PASSWORD` | да | пароль БД приложения |
| `MAX_BOT_TOKEN` | да | токен бота MAX (подпись `initData`, отправка пушей) |
| `MAX_BOT_NAME` | нет | имя бота для ссылок и QR, по умолчанию `t117_hakaton_max_bot` |
| `DADATA_TOKEN` | да | API-ключ DaData |
| `UK_API_TOKEN` | нет | токен обмена с системой УК, по умолчанию `demo-uk-token` |
| `UK_JWT_PRIVATE_KEY` | нет | EC P-256 ключ (PEM) для JWT диспетчеров; пусто — генерируется при старте |
| `GISGKH_DISABLED` | нет | непустое значение отключает запросы в ГИС ЖКХ |

Секреты в репозиторий не коммитятся: `.env` в `.gitignore`, в CI используются GitHub Secrets.

## Деплой

GitHub Actions → **Deploy** → Run workflow (`.github/workflows/deploy.yml`): собирает фронтенд
и Docker-образы, пушит в GHCR, на сервере выполняет `docker compose pull && docker compose up -d`
и smoke-тест. SHA выкаченного коммита сохраняется на сервере в `/opt/municipal/.deployed`.

## Документация

- `docs/frontend-api-contract.md` — контракт API для фронтенда
- `docs/uk-integration.md` — как подключить реальную систему УК
- `docs/ml/` — датасет, обучение и оценка модели дедупликации
- `ml/dedup/README.md` — воспроизведение обучения
