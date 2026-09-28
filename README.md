# Дом.Пульс — мини-приложение MAX для заявок в управляющую компанию

Жители дома сообщают об авариях (нет воды, света, сломан лифт) прямо в MAX. Одинаковые заявки
соседей автоматически склеиваются в одну, УК видит сгруппированный поток, жители получают
уведомления о каждом изменении статуса.

- Мини-приложение: бот [`t117_hakaton_max_bot`](https://max.ru/t117_hakaton_max_bot) → «Открыть приложение»
- API: https://nerionapp.ru — спецификация [`backend/openapi.yaml`](backend/openapi.yaml), проверки [`DATA-API.yaml`](DATA-API.yaml)

## Основной сценарий

1. Житель открывает мини-приложение и привязывает дом: адрес, геолокация или QR-наклейка в
   подъезде. УК дома определяется по данным ГИС ЖКХ.
2. Описывает аварию и прикладывает фото. Категорию и ответственного (УК или муниципалитет)
   подсказывает ML-классификатор.
3. Сосед пишет о том же своими словами — ML-дедупликация присоединяет его к существующей
   заявке, счётчик затронутых жителей растёт. У диспетчера одна карточка вместо десяти.
4. Диспетчер УК видит очередь по числу затронутых жителей и критичности и ведёт заявку:
   принята → в работе → выполнено. Жители получают уведомление на каждом шаге.
5. Жители подтверждают «Починили» — заявка закрывается, когда подтвердили не меньше двух
   и не меньше половины подписанных.

Ещё: ручное объединение заявок и объявления о плановых работах от УК,
[интеграция с 1С УК](docs/uk-1c-integration.md) — лента изменений заявок и обратная смена статусов
(отдельный контракт [`docs/uk-1c-openapi.yaml`](docs/uk-1c-openapi.yaml)).

**Возможности MAX сверх минимума:** уведомления от бота со ссылкой прямо на заявку; шеринг
заявки в домовой чат (`shareMaxContent`) — соседи присоединяются в одно нажатие; QR-наклейка
с `start_param` открывает приложение уже привязанным к дому; геолокация; кабинет УК внутри
того же мини-приложения.

## Архитектура

```
MAX (мини-приложение, frontend/)
        │  Authorization: tma <initData>
        ▼
api (Go, backend/) ──── PostgreSQL
  │   │   │   └── laya    (Node.js, классификатор категорий)
  │   │   └────── dedup   (Python, ML-дедупликация)
  │   └────────── mock-uk (Go, имитация системы УК) ── PostgreSQL
  └── внешние API: MAX Bot API, DaData, ГИС ЖКХ
```

| Сервис | Каталог | Стек |
|---|---|---|
| `api` | `backend/` | Go 1.26, gin, pgx, go-jet, OpenAPI 3.0 |
| `laya` | `backend/laya/` | Node.js 20, ONNX-модель `convaiinnovations/laya` |
| `dedup` | `ml/dedup/` | Python 3.10, onnxruntime, cross-encoder на базе `sergeyzh/BERTA` |
| `mock-uk` | `mock/uk/` | Go 1.26, веб-кабинет диспетчера `/lk` |
| `frontend` | `frontend/` | React 19, Vite, `@maxhub/max-ui` |

Зависимости с версиями — [`requirements.txt`](requirements.txt) (плюс `go.sum`,
`package-lock.json`). Все библиотеки и модели — open-source (MIT, Apache-2.0, BSD-3-Clause,
PostgreSQL License).

### Авторизация

| Кто | Как |
|---|---|
| Житель | `Authorization: tma <initData>` — подпись бота MAX (HMAC-SHA256), срок 24 ч |
| Диспетчер УК | ИНН и пароль → `POST /api/auth/esia-mock` → JWT (ES256), `Authorization: Bearer <jwt>`. bcrypt, после 5 неудачных попыток за 15 минут — блокировка |
| 1С УК | `Authorization: Bearer uk_live_…` — ключ из кабинета УК, хранится хэшем, до 10 ключей на УК, 10 запросов/с |
| Система УК ↔ api | статический токен `UK_API_TOKEN` |

Ручки `/api/uk/*` и смена статуса — только диспетчер своей УК; создание, присоединение и
подтверждение заявок — только житель.

### Внешние сервисы

Не поднимаются в Docker. Для полного сценария в MAX нужны все; локальная проверка API проходит без них.

| Сервис | Для чего | Без ключа |
|---|---|---|
| MAX Bot API | уведомления, подпись `initData` (`MAX_BOT_TOKEN`) | любая строка: `initData` подписывается ей же через `scripts/initdata.py`, уведомления не доходят |
| DaData | подсказки адреса, ФИАС-id дома (`DADATA_TOKEN`) | не работает привязка по адресу; по QR (`start_param=h_1`) работает |
| ГИС ЖКХ | реальная УК дома (ключ не нужен) | `GISGKH_DISABLED=1` — УК берётся из mock-uk |

Системы УК и ЕСИА закрыты, поэтому в MVP их заменяют `mock-uk` и `esia-mock`. Обмен с mock-uk
идёт по контракту, который потребуется от реальной УК ([`docs/uk-integration.md`](docs/uk-integration.md)).

## Запуск

Нужны Docker с Compose, `make` и `gh` с доступом к репозиторию (веса моделей — в GitHub Releases).

```bash
make up
```

Первый запуск создаст `.env` и остановится: заполните `POSTGRES_PASSWORD` и `MAX_BOT_TOKEN`
и повторите. Дальше `make up` скачает веса моделей и выполнит `docker compose up -d --build`.
Сборка с нуля — около 3 минут, ML-сервисы готовы ещё через 1–2 минуты (`docker compose ps` → `healthy`).

Без `make`:

```bash
cp .env.example .env
gh release download dedup-ftM2-berta-s16 -R rnegic/municipal -p bundle.tar.gz -D ml/dedup
gh release download laya-onnx-55cf4c4    -R rnegic/municipal -p model.tar.gz  -D backend/laya
docker compose up -d --build
```

Фронтенд работает внутри MAX и в compose не входит; локально — `cd frontend && npm ci && npm run dev`.

| `make` | `docker compose` | |
|---|---|---|
| `make stop` | `stop` | остановить, данные сохраняются |
| `make down` | `down` | удалить контейнеры, данные сохраняются |
| `make reset` | `down -v` | удалить и данные; при старте демо-данные создаются заново |
| `make logs` | `logs -f` | логи |

### Порты

| Порт | Сервис |
|---|---|
| `127.0.0.1:8080` | `api`, `GET /health` |
| `127.0.0.1:8081` | `mock-uk`, кабинет http://localhost:8081/lk/login |
| внутри сети compose | `laya:8090`, `dedup:8090`, PostgreSQL `db` и `mock-uk-db` |

### Переменные окружения

| Переменная | Обязательна | Описание |
|---|---|---|
| `POSTGRES_PASSWORD` | да | пароль БД |
| `MAX_BOT_TOKEN` | да | токен бота MAX; для локальной проверки — любая строка |
| `MAX_BOT_NAME` | нет | имя бота, по умолчанию `t117_hakaton_max_bot` |
| `DADATA_TOKEN` | для привязки по адресу | ключ DaData |
| `UK_API_TOKEN` | нет | токен обмена с системой УК, по умолчанию `demo-uk-token` |
| `UK_JWT_PRIVATE_KEY` | нет | EC P-256 ключ (PEM) для JWT; пусто — генерируется при старте |
| `GISGKH_DISABLED` | нет | отключает запросы в ГИС ЖКХ |

Тонкие настройки ML и mock-uk — в [`docker-compose.yml`](docker-compose.yml). `.env` не коммитится.

## Проверка

### Тестовые данные

| Роль | Вход |
|---|---|
| Диспетчер УК «Наш Дом» | ИНН `1655000003`, пароль `admin2026` — «Вход для сотрудников УК (ЕСИА)» в приложении |
| Житель | любой пользователь MAX; для API — заголовок из `scripts/initdata.py` |

При старте создаются демо-УК «Наш Дом», дом `h_1` «Казань, ул. Баумана, д. 7/10», три жителя
и три заявки. Запросы и ожидаемые ответы — [`test-data.json`](test-data.json), все
обязательные проверки API — [`DATA-API.yaml`](DATA-API.yaml).

### Сценарий через API

Локальный стенд с `MAX_BOT_TOKEN=test-bot-token`:

```bash
B=http://localhost:8080
anna=$(python3 scripts/initdata.py --token test-bot-token --user 900000101 --name Анна)
boris=$(python3 scripts/initdata.py --token test-bot-token --user 900000102 --name Борис)
json='Content-Type: application/json'

curl -s $B/api/me -H "Authorization: $anna"          # 200, house.id = h_1 (привязка по start_param, как с QR)

curl -s $B/api/incidents -H "Authorization: $anna" -H "$json" \
  -d '{"title":"Нет горячей воды","description":"В третьем подъезде нет горячей воды с утра","category":"WATER_HEAT","entrance":"3","photoUrls":[]}'
                                                      # 201, id = inc_N, affectedCount = 1
curl -s $B/api/incidents -H "Authorization: $boris" -H "$json" \
  -d '{"title":"Горячая вода","description":"Горячую воду отключили, из крана идёт только холодная","category":"WATER_HEAT","entrance":"3","photoUrls":[]}'
                                                      # 200, тот же inc_N, affectedCount = 2

T=$(curl -s $B/api/auth/esia-mock -H "$json" -d '{"inn":"1655000003","password":"admin2026"}' \
  | python3 -c 'import json,sys;print(json.load(sys.stdin)["token"])')
for s in accepted in_progress verifying; do
  curl -s -X PATCH $B/api/incidents/inc_N/status -H "Authorization: Bearer $T" -H "$json" -d "{\"status\":\"$s\"}"
done                                                  # 200, status = $s

curl -s -X POST $B/api/incidents/inc_N/confirm -H "Authorization: $anna"    # 200, verifying
curl -s -X POST $B/api/incidents/inc_N/confirm -H "Authorization: $boris"   # 200, done
```

В MAX: открыть бота, привязать дом «Казань, Баумана 7/10», создать заявку; со второго
аккаунта написать о том же — заявка станет общей, счётчик жителей вырастет.

### Поведение при ошибках

| Ситуация | Ответ |
|---|---|
| Нет авторизации или неверная подпись | `401` |
| Житель вызывает ручку диспетчера | `403` |
| Неверный пароль диспетчера | `403`, после 5 попыток за 15 минут — `429` |
| Некорректные поля заявки | `400 validation_failed` |
| Проблема не в зоне УК | `422 business_rule_failed` с подсказкой, куда обратиться |
| Несуществующая заявка | `404` |
| Классификатор недоступен | `200`, `category = null` — категория выбирается вручную |
| Дедупликация недоступна | склейка по правилу: тот же дом, категория и подъезд за 2 часа |
| mock-УК недоступна | заявка сохраняется, регистрация в УК повторяется в фоне |

## Работа с данными

- `db` (том `pgdata`): УК, дома, пользователи, заявки, подписки и подтверждения, фото (до 10 МБ),
  плановые работы, очередь уведомлений, хэши ключей 1С, журнал входов. Схема и демо-данные
  применяются при старте `api`.
- `mock-uk-db` (том `mockuk-pgdata`): данные имитации системы УК.
- О жителе храним только имя, аватар и id пользователя MAX. Пароли — bcrypt, ключи 1С — хэш.
  Фото без заявки удаляются через 24 часа. ML-модели работают локально в контейнерах, в сторонние ML-сервисы текст заявок не уходит.

## Известные ограничения

- ЕСИА и системы УК имитируются. Доступ в кабинет УК выдаётся вручную (в демо — только «Наш Дом»);
  срок лицензии берётся из нашей БД, а не из реестра ГИС ЖКХ.
- Лимиты запросов и защита от перебора пароля хранятся в памяти — рассчитаны на один экземпляр `api`.
- Без `DADATA_TOKEN` привязка по адресу отвечает `500` — используйте QR.
- Уведомления доставляются только с настоящим токеном бота MAX.
- Пока у УК есть активный ключ 1С, её заявки не передаются в mock-УК.
- `laya` требует до 3 ГБ памяти, `dedup` — до 1,5 ГБ; веса моделей — в GitHub Releases.

## Документация

- [`docs/uk-integration.md`](docs/uk-integration.md) — подключение системы УК
- [`docs/uk-1c-integration.md`](docs/uk-1c-integration.md) — подключение 1С, контракт — [`docs/uk-1c-openapi.yaml`](docs/uk-1c-openapi.yaml)
- [`ml/dedup/README.md`](ml/dedup/README.md) — воспроизведение обучения
