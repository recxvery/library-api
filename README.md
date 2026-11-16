# Library API

REST-сервис на Go для небольшой домашней библиотеки. Позволяет добавлять книги, смотреть список и карточку книги, отмечать книгу прочитанной и удалять её.

Проект учебный, но собран как обычный backend-сервис: PostgreSQL, миграции, слои handler/repository, структурные логи, middleware для логирования, восстановления после паники и graceful shutdown. Всё поднимается одной командой через Docker Compose.

## Что умеет

- создавать книгу с проверкой обязательных полей;
- отдавать список книг с фильтрами по автору и статусу прочтения;
- отдавать одну книгу по идентификатору;
- отмечать книгу прочитанной с сохранением даты прочтения;
- удалять книгу и возвращать удалённую запись;
- автоматически применять SQL-миграции при старте;
- писать логи в консоль и в файл;
- корректно завершаться по `SIGINT`/`SIGTERM`.

## Технологии

- Go 1.25
- [gorilla/mux](https://github.com/gorilla/mux) — маршрутизация
- [pgx/v5](https://github.com/jackc/pgx) и `pgxpool` — работа с PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) — миграции
- [zap](https://github.com/uber-go/zap) — логирование
- PostgreSQL 17
- Docker и Docker Compose

## Структура проекта

```
.
├── cmd/server/           точка входа: подключение к БД, миграции, запуск сервера
├── internal/
│   ├── core/             логгер и общие ошибки
│   ├── db/               подключение к PostgreSQL, модели и репозиторий книг
│   └── http/             HTTP-хендлеры, DTO, middleware, сервер и тесты
├── migrations/           SQL-миграции
├── logs/                 файлы логов
├── docker-compose.yaml
├── Dockerfile
└── Makefile
```

## Быстрый старт через Docker

1. Создайте файл `.env` на основе `.env.example`.

   ```bash
   cp .env.example .env
   ```

2. Заполните переменные. Пример:

   ```dotenv
   POSTGRES_USER=library
   POSTGRES_PASS=library
   POSTGRES_DB=library
   CONN_STRING=postgres://library:library@db:5432/library
   ```

   В `CONN_STRING` хост должен быть `db` — это имя сервиса базы внутри Docker-сети. Логин, пароль и имя базы должны совпадать с `POSTGRES_*`.

3. Поднимите сервисы.

   ```bash
   docker compose up --build
   ```

   Или через Makefile:

   ```bash
   make service-up-rebuild
   ```

4. Сервис будет доступен по адресу `http://localhost:8081`. Миграции применяются автоматически при запуске приложения.

Остановить контейнеры:

```bash
docker compose down
```

Только база данных, без приложения:

```bash
make service-db-up
```

## Запуск без Docker

Понадобятся Go и локальный PostgreSQL.

1. Создайте базу данных.
2. Задайте строку подключения:

   ```bash
   export CONN_STRING="postgres://user:password@localhost:5432/library"
   ```

   На Windows PowerShell:

   ```powershell
   $env:CONN_STRING = "postgres://user:password@localhost:5432/library"
   ```

3. Запустите сервис:

   ```bash
   go run ./cmd/server
   ```

Параметры подключения в `CONN_STRING` лучше не указывать: для миграций приложение само добавляет `?sslmode=disable`. Пул соединений использует строку как есть.

## Переменные окружения

| Переменная | Где используется | Назначение |
|---|---|---|
| `CONN_STRING` | приложение | строка подключения к PostgreSQL |
| `POSTGRES_USER` | контейнер `db` | пользователь базы |
| `POSTGRES_PASS` | контейнер `db` | пароль пользователя |
| `POSTGRES_DB` | контейнер `db` | имя базы |

## API

Базовый адрес: `http://localhost:8081`.

| Метод | Путь | Что делает |
|---|---|---|
| `POST` | `/books` | добавляет книгу |
| `GET` | `/books` | возвращает список книг |
| `GET` | `/books/{id}` | возвращает книгу по id |
| `PATCH` | `/books/{id}` | отмечает книгу прочитанной |
| `DELETE` | `/books/{id}` | удаляет книгу и возвращает её данные |

### Добавить книгу

```bash
curl -X POST http://localhost:8081/books \
  -H "Content-Type: application/json" \
  -d '{
    "author": "J.R.R. Tolkien",
    "title": "The Hobbit",
    "description": "A story about a journey",
    "pages": 310,
    "year": 1937
  }'
```

Правила валидации:

- `author` и `title` не могут быть пустыми;
- `pages` должно быть больше 30;
- `year` должен быть больше 100;
- `description` можно не передавать.

Ответ `201 Created`:

```json
{
  "ID": 1,
  "Title": "The Hobbit",
  "Author": "J.R.R. Tolkien",
  "Description": "A story about a journey",
  "Year": 1937,
  "PagesCount": 310,
  "IsRead": false,
  "AddedAt": "2026-05-12T10:30:00Z",
  "ReadAt": null
}
```

### Получить список книг

```bash
curl http://localhost:8081/books
```

Фильтры передаются query-параметрами:

```bash
curl "http://localhost:8081/books?author=George%20Orwell"
curl "http://localhost:8081/books?read=true"
curl "http://localhost:8081/books?author=George%20Orwell&read=false"
```

`author` сравнивается точно, `read` принимает `true` или `false`. Ответ — массив книг в том же формате, что и выше.

### Получить книгу по id

```bash
curl http://localhost:8081/books/1
```

### Отметить книгу прочитанной

```bash
curl -X PATCH http://localhost:8081/books/1
```

У книги выставится `IsRead: true` и заполнится `ReadAt`.

### Удалить книгу

```bash
curl -X DELETE http://localhost:8081/books/1
```

Возвращается удалённая книга.

### Ошибки

Ответы с ошибкой приходят в виде JSON:

```json
{
  "Message": "no rows in result set",
  "Time": "2026-05-12T10:30:00.123456+03:00"
}
```

Коды ответа:

| Код | Когда возвращается |
|---|---|
| `400 Bad Request` | некорректный id, битый JSON, неверное значение `read` |
| `404 Not Found` | книги с таким id нет |
| `500 Internal Server Error` | ошибка базы или другая непредвиденная ошибка |
| `503 Service Unavailable` | сервер завершает работу и не принимает новые запросы |

## База данных

Таблица `books` создаётся миграциями:

| Поле | Тип | Описание |
|---|---|---|
| `id` | `SERIAL` | первичный ключ |
| `book_title` | `VARCHAR(1000)` | название |
| `book_author` | `VARCHAR(200)` | автор |
| `publish_year` | `INTEGER` | год издания |
| `pages_count` | `INTEGER` | количество страниц |
| `book_read` | `BOOLEAN` | прочитана ли книга |
| `added_at` | `TIMESTAMP` | дата добавления |
| `read_at` | `TIMESTAMP` | дата прочтения |
| `book_description` | `VARCHAR(1000)` | описание |

Миграции лежат в `migrations/` и применяются при каждом старте приложения. Если новых миграций нет, сервис продолжает работу без ошибок.

## Логи

Логи пишутся одновременно в stdout и в файл `logs/app logs`. Формат — консольный, с временем, уровнем, сообщением и caller'ом. Для каждого HTTP-запроса пишется строка с методом, путём, статусом, временем выполнения, IP и User-Agent.

## Тесты

```bash
go test ./...
```

Подробный вывод:

```bash
go test -v ./internal/http
```

Покрытие:

```bash
go test -cover ./...
```

Тесты не требуют базы данных: HTTP-хендлеры проверяются через `httptest` и подменённый репозиторий. Проверяются успешные сценарии, ошибки валидации, некорректные id и query-параметры, ответы `400`, `404`, `500`, а также middleware логирования, восстановления после паники и отклонения запросов при завершении работы.

## Makefile

| Команда | Действие |
|---|---|
| `make service-db-up` | поднять только PostgreSQL |
| `make service-up` | поднять все сервисы |
| `make service-up-rebuild` | пересобрать и поднять все сервисы |
| `make service-down` | остановить и удалить контейнеры |

## Известные ограничения

- нет пагинации и сортировки списка книг;
- фильтр по автору работает по точному совпадению;
- нет аутентификации и авторизации;
- нет отдельных эндпоинтов для проверки состояния сервиса;
- при ошибке валидации в `POST /books` клиенту сейчас не возвращается описание ошибки — она только пишется в лог;
- `added_at` при создании книги заполняется нулевым значением, потому что дата не устанавливается в хендлере;
- порт `8081` задан в коде.
