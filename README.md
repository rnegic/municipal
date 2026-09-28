# УК-инциденты — мини-приложение MAX для заявок в управляющую компанию

Жители дома сообщают об авариях (нет воды, свет, лифт) прямо в мессенджере MAX. Одинаковые
заявки от соседей автоматически склеиваются в одну, УК получает сгруппированный поток, а жители —
пуш-уведомления о каждом изменении статуса.

- Мини-приложение: бот [`t117_hakaton_max_bot`](https://max.ru/t117_hakaton_max_bot) в MAX → «Открыть приложение»
- API: https://nerionapp.ru (спецификация — [`backend/openapi.yaml`](backend/openapi.yaml), проверки — [`DATA-API.yaml`](DATA-API.yaml))

## Основной пользовательский сценарий

1. Житель открывает бота в MAX и запускает мини-приложение.
2. Привязывает дом: адрес с подсказками, геолокация или QR-наклейка в подъезде. УК дома
   определяется автоматически по данным ГИС ЖКХ.
3. «Сообщить о проблеме»: описывает аварию, прикладывает фото. Категорию и ответственного
   (УК или муниципалитет) подсказывает ML-классификатор.
4. Сосед пишет о той же аварии своими словами — ML-дедупликация присоединяет его к существующей
   заявке, счётчик «затронуто жителей» растёт. У диспетчера — одна карточка вместо десяти.
5. Диспетчер УК в кабинете видит очередь, отсортированную по числу затронутых жителей и
   критичности, и ведёт заявку: принята → в работе → выполнено.
6. Жители получают пуш на каждом шаге. После «выполнено» они подтверждают «Починили» — заявка
   закрывается, когда подтвердили не меньше двух жителей и не меньше половины подписанных.

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
- Кабинет УК в мини-приложении: доска заявок своих домов, смена статусов, объединение, объявление плановых работ.
- Интеграция с 1С УК: ключ выпускается в кабинете, 1С забирает ленту изменений заявок по курсору
  и отправляет смену статуса обратно — см. [гайд подключения 1С](docs/uk-1c-integration.md).

### Возможности MAX сверх минимальных требований

- **Пуш-уведомления от бота** со ссылкой, открывающей мини-приложение сразу на нужной заявке.
- **Шеринг в чаты MAX** (`shareMaxContent`): житель отправляет заявку в домовой чат, соседи
  одним нажатием присоединяются к ней.
- **Deep-link со `start_param`**: QR-наклейка в подъезде открывает мини-приложение уже
  привязанным к дому — без ввода адреса.
- **Геолокация** для определения дома.
- **Кабинет УК внутри MAX**: диспетчер работает в том же мини-приложении, а не в отдельной системе.

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
| `api` | `backend/` | Go 1.26, gin, pgx, go-jet, OpenAPI ([`backend/openapi.yaml`](backend/openapi.yaml)) |
| `laya` | `backend/laya/` | Node.js 20, ONNX-модель `convaiinnovations/laya` |
| `dedup` | `ml/dedup/` | Python 3.10, onnxruntime, cross-encoder на базе `sergeyzh/BERTA` |
| `mock-uk` | `mock/uk/` | Go 1.26 — имитация системы УК и веб-кабинет диспетчера |
| `frontend` | `frontend/` | React 19, Vite, `@maxhub/max-ui` |

Список всех зависимостей с лицензиями — [`requirements.txt`](requirements.txt), версии
зафиксированы также в `go.sum`, `package-lock.json` и `ml/dedup/requirements-serve.txt`.
Все используемые библиотеки и модели распространяются под open-source лицензиями
(MIT, Apache-2.0, BSD-3-Clause, PostgreSQL License).

## Техническое ограничение MVP

Реальные базы Госуслуг (ЕСИА) и внутренние системы УК закрыты, поэтому для демонстрации мы
развернули собственный mock-сервер УК и кабинет диспетчера (канбан-доску), куда падают
сгруппированные заявки. Обмен с ним идёт по тому же REST-контракту, что потребуется от реальной
УК ([`docs/uk-integration.md`](docs/uk-integration.md)) — для подключения настоящей системы меняется только адаптер
`backend/internal/ukclient`.

## Авторизация

| Кто | Как | Где в коде |
|---|---|---|
| Житель | Заголовок `Authorization: tma <initData>`. MAX подписывает `initData` ключом бота, бэкенд проверяет HMAC-SHA256 подпись и срок (24 ч). Пользователь создаётся при первом запросе. | `backend/internal/domain` (`ValidateInitData`), `backend/internal/transport/http/middleware.go` |
| Диспетчер УК | Вход по ИНН организации и паролю (`POST /api/auth/esia-mock` — имитация входа через ЕСИА). Пароль — bcrypt, 5 неудачных попыток за 15 минут — блокировка. Выдаётся JWT (ES256) с ролью `uk_dispatcher` и id УК, дальше `Authorization: Bearer <jwt>`. | `backend/internal/ukauth`, `backend/internal/service/ukadmin.go` |
| 1С УК → api | `Authorization: Bearer uk_live_…`. Ключ выпускается диспетчером в кабинете УК (показывается один раз, у нас хранится только хэш), не больше 10 активных ключей на УК, 10 запросов/с на ключ. Ключ действует как диспетчер своей УК; при недействующей лицензии УК в ГИС ЖКХ — `401`. Подробно — [`docs/uk-1c-integration.md`](docs/uk-1c-integration.md). | `backend/internal/ukauth`, `backend/internal/service` |
| Система УК ↔ api | Статический bearer-токен `UK_API_TOKEN` в обе стороны. | `backend/internal/ukclient`, `mock/uk/internal/api` |

Роли проверяются в middleware: ручки `/api/uk/*` и смена статуса — только диспетчер,
создание/присоединение/подтверждение заявок — только житель; диспетчер видит только дома своей УК.

## Внешние API

Эти сервисы нельзя поднять в Docker — они нужны для полного сценария в MAX, но локальная
проверка API проходит и без них (см. «Проверка»).

| API | Для чего | Ключ | Без ключа |
|---|---|---|---|
| MAX Bot API (`platform-api2.max.ru`) | пуш-уведомления жителям, проверка `initData` | `MAX_BOT_TOKEN` | подойдёт любая строка: `initData` подписывается ей же (`scripts/initdata.py`), пуши не доходят и остаются в очереди повторов |
| DaData (`suggestions.dadata.ru`) | подсказки адреса, геокодирование, ФИАС-id дома | `DADATA_TOKEN` | не работают подсказки и привязка по адресу; привязка по QR (`start_param=h_1`) работает |
| ГИС ЖКХ (`dom.gosuslugi.ru`, публичный поиск по адресу) | реальная УК дома, проверка что дом многоквартирный | не нужен | отключается `GISGKH_DISABLED=1`, УК берётся из mock-uk |
| Система УК (в MVP — `mock-uk`, поднимается в compose) | регистрация заявок, синхронизация статусов | `UK_API_TOKEN` | по умолчанию `demo-uk-token` |

## Запуск

Нужны Docker с Compose и доступ к репозиторию (для весов ML-моделей из GitHub Releases).

```bash
cp .env.example .env    # заполнить POSTGRES_PASSWORD и MAX_BOT_TOKEN, см. ниже
gh release download dedup-ftM2-berta-s16 -p bundle.tar.gz -D ml/dedup      # веса дедупликации, 480 МБ
gh release download laya-onnx-55cf4c4    -p model.tar.gz  -D backend/laya  # ONNX-классификатор, 810 МБ
docker compose up -d --build
```

Без `gh` те же файлы скачиваются со страницы Releases репозитория. Целостность весов
проверяется при сборке по sha256 (`DEDUP_SHA256`, `LAYA_SHA256` в `docker-compose.yml`).
Сборка всех образов с нуля — около 3 минут (без учёта загрузки базовых образов); `laya` и
`dedup` готовы к работе ещё через 1–2 минуты после старта (`docker compose ps` → `healthy`).

Модель `laya` в релизе — результат этапа `export` в [`backend/laya/Dockerfile`](backend/laya/Dockerfile)
(скачивание чекпойнта с HuggingFace и экспорт в ONNX, ~6 минут). Пересобрать её самостоятельно:
`docker build --target model -o out backend/laya && tar -czf backend/laya/model.tar.gz -C out .`

Готовые образы команды лежат в GHCR (`ghcr.io/rnegic/municipal-*`, приватные) — с доступом
вместо сборки можно `docker compose pull && docker compose up -d`.

Фронтенд в compose не входит: это мини-приложение, оно работает внутри MAX (в проде
собирается в CI и отдаётся nginx). Локально: `cd frontend && npm ci && npm run dev`.

### Порты

| Порт | Сервис | Что там |
|---|---|---|
| `127.0.0.1:8080` | `api` | REST API, `GET /health` |
| `127.0.0.1:8081` | `mock-uk` | кабинет диспетчера mock-УК: http://localhost:8081/lk/login |
| внутри сети compose | `laya:8090`, `dedup:8090` | ML-сервисы, наружу не публикуются |
| внутри сети compose | `db:5432`, `mock-uk-db:5432` | PostgreSQL 16 |
| `5173` | frontend (`npm run dev`) | только при локальной разработке фронтенда |

### Переменные окружения (`.env`)

| Переменная | Обязательна | Описание |
|---|---|---|
| `POSTGRES_PASSWORD` | да | пароль БД приложения |
| `MAX_BOT_TOKEN` | да | токен бота MAX (подпись `initData`, отправка пушей); для локальной проверки — любая строка |
| `MAX_BOT_NAME` | нет | имя бота для ссылок и QR, по умолчанию `t117_hakaton_max_bot` |
| `DADATA_TOKEN` | для привязки по адресу | API-ключ DaData |
| `UK_API_TOKEN` | нет | токен обмена с системой УК, по умолчанию `demo-uk-token` |
| `UK_JWT_PRIVATE_KEY` | нет | EC P-256 ключ (PEM) для JWT диспетчеров; пусто — генерируется при старте |
| `GISGKH_DISABLED` | нет | непустое значение отключает запросы в ГИС ЖКХ |

Необязательные настройки ML и mock-uk (`DEDUP_THREADS`, `DEDUP_TIMEOUT`, `DEDUP_SHA256`,
`LAYA_THREADS`, `LAYA_CATEGORY_THRESHOLD`, `LAYA_SHA256`, `MOCK_FAILURE_RATE`) — значения по умолчанию
в [`docker-compose.yml`](docker-compose.yml).

Секреты в репозиторий не коммитятся: `.env` в `.gitignore`, в CI используются GitHub Secrets.

### Остановка и повторный запуск

```bash
docker compose stop                # остановить, данные сохраняются
docker compose up -d               # запустить снова
docker compose down                # удалить контейнеры, тома с БД остаются
docker compose down -v             # удалить и данные — при следующем старте схема и демо-данные создаются заново
docker compose logs -f api         # логи
```

## Тестовые учётные записи и данные

| Роль | Вход | Где |
|---|---|---|
| Диспетчер УК «Наш Дом» | ИНН `1655000003`, пароль `admin2026` | кабинет УК в мини-приложении, `POST /api/auth/esia-mock` |
| Диспетчер mock-УК | `dispatcher` / `demo1234` | http://localhost:8081/lk/login |
| Житель | любой пользователь MAX; для API — заголовок из `scripts/initdata.py` | мини-приложение, `Authorization: tma …` |

При первом старте `api` создаёт схему и демо-данные ([`backend/internal/repository/seed.sql`](backend/internal/repository/seed.sql)):
УК «Наш Дом» с действующей лицензией, демо-дом `h_1` «Казань, ул. Баумана, д. 7/10»,
три жителя и три заявки. mock-uk заводит ту же организацию и дом
([`mock/uk/internal/store/seed.sql`](mock/uk/internal/store/seed.sql)). Сид идемпотентен —
повторный старт ничего не дублирует.

Тестовые данные для проверки — [`test-data.json`](test-data.json): жители, пара заявок-дубликатов,
заявки для ручной склейки, заведомо ошибочные запросы и ожидаемые ответы.

## Проверка

Все обязательные проверки API с ролями, запросами и ожидаемыми ответами — в
[`DATA-API.yaml`](DATA-API.yaml). Ниже тот же сценарий вручную на локальном стенде
(`MAX_BOT_TOKEN=test-bot-token` в `.env`):

```bash
B=http://localhost:8080
anna=$(python3 scripts/initdata.py --token test-bot-token --user 900000101 --name Анна)
boris=$(python3 scripts/initdata.py --token test-bot-token --user 900000102 --name Борис)

# 1. Житель привязан к демо-дому через start_param=h_1 (как по QR-наклейке)
curl -s $B/api/me -H "Authorization: $anna"                       # 200, house.id = h_1

# 2. Классификатор определяет категорию по тексту
curl -s $B/api/incidents/analyze -H "Authorization: $anna" -H 'Content-Type: application/json' \
  -d '{"description":"В третьем подъезде нет горячей воды с утра"}'   # 200, category = WATER_HEAT

# 3. Анна создаёт заявку
curl -s $B/api/incidents -H "Authorization: $anna" -H 'Content-Type: application/json' \
  -d '{"title":"Нет горячей воды","description":"В третьем подъезде нет горячей воды с утра","category":"WATER_HEAT","entrance":"3","photoUrls":[]}'
                                                                     # 201, id = inc_N, affectedCount = 1

# 4. Борис пишет о том же другими словами — заявки склеиваются
curl -s $B/api/incidents -H "Authorization: $boris" -H 'Content-Type: application/json' \
  -d '{"title":"Горячая вода","description":"Горячую воду отключили, из крана идёт только холодная","category":"WATER_HEAT","entrance":"3","photoUrls":[]}'
                                                                     # 200, тот же id, affectedCount = 2

# 5. Диспетчер входит и ведёт заявку
T=$(curl -s $B/api/auth/esia-mock -H 'Content-Type: application/json' \
  -d '{"inn":"1655000003","password":"admin2026"}' | python3 -c 'import json,sys;print(json.load(sys.stdin)["token"])')
curl -s "$B/api/uk/queue" -H "Authorization: Bearer $T"            # 200, заявка первой, affectedCount = 2
for s in accepted in_progress verifying; do
  curl -s -X PATCH $B/api/incidents/inc_N/status -H "Authorization: Bearer $T" \
    -H 'Content-Type: application/json' -d "{\"status\":\"$s\"}"    # 200, status = $s
done

# 6. Жители подтверждают — заявка закрыта
curl -s -X POST $B/api/incidents/inc_N/confirm -H "Authorization: $anna"    # 200, status = verifying
curl -s -X POST $B/api/incidents/inc_N/confirm -H "Authorization: $boris"   # 200, status = done
```

Ожидаемое поведение на ошибках:

| Ситуация | Ответ |
|---|---|
| Нет заголовка авторизации или неверная подпись `initData` | `401 unauthorized` |
| Житель вызывает ручку диспетчера (`/api/uk/*`, смена статуса) | `403 forbidden` |
| Неверный пароль диспетчера | `403`; после 5 попыток за 15 минут — `429` |
| Описание короче 10 символов, неизвестная категория | `400 validation_failed` |
| Проблема не в зоне УК (яма на дороге) | `422 business_rule_failed` с подсказкой, куда обратиться |
| Несуществующая заявка | `404 not_found` |
| Классификатор недоступен или не уверен | `200`, `category = null` — выбор категории вручную |
| Сервис дедупликации недоступен | склейка по правилу «тот же дом, категория и подъезд за 2 часа» |
| mock-УК недоступна | заявка сохраняется, регистрация в УК повторяется в фоне |

В MAX тот же сценарий: открыть бота, пройти онбординг, привязать дом «Казань, Баумана 7/10»,
создать заявку; со второго аккаунта — написать о том же: заявка окажется общей, счётчик
затронутых жителей вырастет. Кабинет УК — «Вход для сотрудников УК (ЕСИА)» на экране онбординга,
ИНН и пароль из таблицы выше.

## Работа с данными

- **PostgreSQL `db`** (том `pgdata`): УК, дома, пользователи, заявки, подписки и подтверждения
  жителей, фото (в `bytea`, до 10 МБ, HEIC конвертируется в JPEG), плановые работы, очередь
  исходящих уведомлений (`outbox_message`), хэши API-ключей 1С и журнал попыток входа.
  Схема — [`backend/internal/repository/schema.sql`](backend/internal/repository/schema.sql),
  применяется при старте `api`.
- **PostgreSQL `mock-uk-db`** (том `mockuk-pgdata`): организации, дома, заявки и диспетчеры
  имитации системы УК.
- Персональные данные жителей — только имя, аватар и id пользователя MAX из `initData`. Пароли — bcrypt
  (mock-УК — sha256), API-ключи хранятся хэшем. Непривязанные к заявке фото удаляются через 24 часа.
- В ML-сервисы уходит только текст заявок, модели работают локально в контейнерах.

## Известные ограничения

- ЕСИА и системы УК закрыты — вход диспетчера и система УК имитируются (`esia-mock`, `mock-uk`).
- Лимит 10 запросов/с на ключ 1С и защита от перебора пароля хранятся в памяти — рассчитаны на
  один экземпляр `api`.
- Без `DADATA_TOKEN` привязка по адресу отвечает `500` — используйте QR (`start_param=h_1`).
- Пуши доставляются только с настоящим токеном бота MAX; при сбоях очередь повторяет отправку.
- Автосклейка ищет дубликаты среди открытых заявок того же дома и категории за последние 2 часа.
- Пока у УК есть активный ключ 1С, её заявки не передаются в mock-УК — источником статусов считается 1С.
- `laya` требует до 3 ГБ памяти, `dedup` — до 1,5 ГБ.
- Веса моделей хранятся в GitHub Releases, а не в git (размер).

## Деплой

GitHub Actions → **Deploy** → Run workflow ([`.github/workflows/deploy.yml`](.github/workflows/deploy.yml)): собирает фронтенд
и Docker-образы, пушит в GHCR, на сервере выполняет `docker compose pull && docker compose up -d`
и smoke-тест. SHA выкаченного коммита сохраняется на сервере в `/opt/municipal/.deployed`.

## Документация

- [`backend/openapi.yaml`](backend/openapi.yaml) — спецификация API (OpenAPI 3.0)
- [`DATA-API.yaml`](DATA-API.yaml) — обязательные проверки API
- [`test-data.json`](test-data.json) — тестовые данные
- [`docs/frontend-api-contract.md`](docs/frontend-api-contract.md) — контракт API для фронтенда
- [`docs/uk-integration.md`](docs/uk-integration.md) — как подключить реальную систему УК
- [`docs/uk-1c-integration.md`](docs/uk-1c-integration.md) — как подключить 1С управляющей компании
- [`docs/ml/`](docs/ml/) — модель дедупликации:
  [датасет](docs/ml/dedup-dataset.md), [обучение](docs/ml/dedup-training.md),
  [оценка](docs/ml/dedup-evaluation.md), [прод](docs/ml/dedup-production.md),
  [отчёт](docs/ml/dedup-report.md)
- [`ml/dedup/README.md`](ml/dedup/README.md) — воспроизведение обучения
