# Контракт интеграции frontend ↔ backend

Этот документ можно передать backend-разработчику. В нём зафиксировано, что уже реализовано на frontend, какие запросы он отправляет и какие ответы ожидает.

## 1. Текущее состояние

Frontend находится в папке `frontend/` и работает с API через модуль `frontend/src/services/api.ts`.

Сейчас полностью согласована загрузка документов:

- `POST /api/v1/documents/upload`;
- формат `multipart/form-data`;
- имя поля с файлом — `file`;
- один файл передаётся одним запросом;
- множественная загрузка реализована на frontend как несколько независимых запросов.

Поисковый интерфейс на frontend готов, но backend пока должен реализовать и зарегистрировать endpoint `GET /api/v1/search`.

## 2. Адрес API

При локальной разработке:

- frontend: `http://localhost:5173`;
- backend: `http://localhost:8080`;
- браузер обращается к `/api/...` на адресе frontend;
- Vite проксирует запросы на `http://localhost:8080`.

Поэтому при обычном локальном запуске frontend отдельная настройка CORS не требуется.

Если frontend будет обращаться напрямую к `http://localhost:8080`, backend должен разрешить origin `http://localhost:5173`, методы `GET`, `POST`, `OPTIONS` и заголовок `Content-Type`.

## 3. Загрузка документа

### Запрос

```http
POST /api/v1/documents/upload
Content-Type: multipart/form-data
```

Поле формы:

| Поле | Тип | Обязательное | Описание |
|---|---|---:|---|
| `file` | File | Да | PDF или DOCX размером до 20 МБ |

Пример проверки через curl:

```bash
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -F "file=@lecture.pdf"
```

### Успешный ответ

HTTP `200 OK`:

```json
{
  "id": "0c6e693c-5f88-4a45-95ac-659f0e93837d",
  "file_name": "lecture.pdf",
  "size": 1048576,
  "status": "text_extracted",
  "message": "file uploaded and text extracted successfully",
  "content_hash": "sha256-value",
  "pages_count": 24,
  "extracted_chars": 48520,
  "text_preview": "Начало извлечённого текста...",
  "duplicate": false
}
```

Frontend обязательно использует следующие поля:

- `id`;
- `file_name`;
- `size`;
- `status`;
- `pages_count`;
- `duplicate`.

Если документ уже загружался, backend может вернуть:

```json
{
  "id": "new-request-id",
  "original_document_id": "existing-document-id",
  "file_name": "lecture.pdf",
  "size": 1048576,
  "status": "duplicate",
  "message": "file already uploaded; extraction skipped",
  "pages_count": 24,
  "duplicate": true
}
```

Frontend отобразит статус «Уже загружен».

### Ошибка валидации

HTTP `400 Bad Request`:

```json
{
  "error": "validation_error",
  "message": "invalid file: only PDF and DOCX files are allowed"
}
```

Поле `message` выводится пользователю, поэтому оно должно содержать понятное описание ошибки.

## 4. Поиск

Backend должен реализовать следующий endpoint.

### Запрос

```http
GET /api/v1/search?q=жизненный%20цикл&page=1&limit=10
Accept: application/json
```

Параметры:

| Параметр | Тип | Обязательный | Описание |
|---|---|---:|---|
| `q` | string | Да | Непустой поисковый запрос |
| `page` | integer | Нет | Номер страницы, начиная с 1; по умолчанию 1 |
| `limit` | integer | Нет | Количество результатов; frontend передаёт 10 |

### Рекомендуемый успешный ответ

HTTP `200 OK`:

```json
{
  "items": [
    {
      "chunk_id": "document-id:chunk-12",
      "file_name": "Программная инженерия.pdf",
      "page": 7,
      "text": "Жизненный цикл программного продукта включает анализ, проектирование, разработку и тестирование.",
      "score": 0.934
    }
  ],
  "total": 24,
  "page": 1,
  "total_pages": 3
}
```

Обязательные поля результата:

| Поле | Тип | Описание |
|---|---|---|
| `chunk_id` | string | Уникальный идентификатор найденного фрагмента |
| `file_name` | string | Исходное имя документа |
| `page` | integer | Номер страницы документа, начиная с 1 |
| `text` | string | Найденный фрагмент без HTML-разметки |
| `score` | number | Релевантность: рекомендуется значение от 0 до 1 |

Frontend самостоятельно подсвечивает слова из запроса. Backend не должен добавлять в `text` теги `<mark>`, `<em>` или другую HTML-разметку.

Для пагинации необходимо возвращать:

- `total` — полное количество найденных фрагментов;
- `page` — текущую страницу;
- `total_pages` — общее количество страниц.

Frontend также умеет прочитать массив результатов без обёртки и поле `results` вместо `items`, но объект, показанный выше, является основным согласованным форматом.

### Пустая выдача

Это не ошибка. Нужно вернуть `200 OK`:

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "total_pages": 1
}
```

Frontend покажет сообщение: «По вашему запросу ничего не найдено. Попробуйте изменить формулировку».

### Ошибки поиска

Некорректный запрос — HTTP `400 Bad Request`:

```json
{
  "error": "validation_error",
  "message": "query parameter q must not be empty"
}
```

Внутренняя ошибка — HTTP `500 Internal Server Error`:

```json
{
  "error": "internal_error",
  "message": "failed to search documents"
}
```

Во всех ошибках желательно сохранять единый формат с полями `error` и `message`.

## 5. Список документов

В исходном задании frontend должен отображать ранее загруженные документы. Сейчас frontend временно сохраняет их метаданные в `localStorage`, потому что соответствующего endpoint на backend нет.

Для полноценной синхронизации желательно реализовать:

```http
GET /api/v1/documents
```

Рекомендуемый ответ:

```json
{
  "items": [
    {
      "id": "0c6e693c-5f88-4a45-95ac-659f0e93837d",
      "file_name": "lecture.pdf",
      "size": 1048576,
      "status": "ready",
      "pages_count": 24,
      "uploaded_at": "2026-07-02T12:00:00Z"
    }
  ],
  "total": 1
}
```

Этот endpoint не блокирует первую интеграцию загрузки и поиска, но нужен для одинакового списка документов в разных браузерах.

## 6. Статусы обработки

Frontend показывает четыре основных состояния:

1. «Загрузка» — браузер отправляет тело запроса.
2. «Индексация» — файл отправлен, frontend ожидает ответ backend.
3. «Готово» — backend вернул успешный ответ.
4. «Ошибка» — получен HTTP-код ошибки или сервер недоступен.

Отдельный endpoint прогресса сейчас не требуется. `POST /upload` должен отвечать только после извлечения текста и завершения необходимой обработки документа.

## 7. Docker и Nginx

Production frontend отдаётся через Nginx. Сейчас его конфигурация направляет `/api/` на:

```text
http://backend:8080
```

В `docker-compose.yml` необходимо выбрать один из вариантов:

1. назвать сервис backend именем `backend`;
2. добавить сервису сетевой alias `backend`;
3. заменить адрес в `frontend/nginx.conf`, если сервис будет называться `app`.

Главное условие: имя в Nginx и имя/alias сервиса в Docker Compose должны совпадать.

## 8. Swagger

Backend-разработчику необходимо описать в OpenAPI 3.0:

- `POST /api/v1/documents/upload`;
- `GET /api/v1/search`;
- `GET /api/v1/documents`, если он будет добавлен;
- схемы успешных ответов;
- ответы `400`, `404` и `500`.

Swagger UI по заданию должен быть доступен по адресу `/docs`.

## 9. Чек-лист backend-разработчика

- [ ] Зарегистрирован маршрут `GET /api/v1/search`.
- [ ] Обрабатываются параметры `q`, `page`, `limit`.
- [ ] В выдаче есть `chunk_id`, `file_name`, `page`, `text`, `score`.
- [ ] Возвращаются `total`, `page`, `total_pages`.
- [ ] Пустой результат возвращается с HTTP 200.
- [ ] Ошибки имеют поля `error` и `message`.
- [ ] Текст результата не содержит HTML-разметки для подсветки.
- [ ] При необходимости реализован `GET /api/v1/documents`.
- [ ] Имя backend-сервиса согласовано с Nginx/Docker Compose.
- [ ] API описано в Swagger и доступно по `/docs`.

## 10. Совместная проверка

После готовности backend необходимо вместе проверить сценарий:

1. Запустить backend на порту 8080.
2. Запустить frontend командой `npm run dev` в папке `frontend/`.
3. Загрузить корректный PDF.
4. Дождаться статуса «Готово».
5. Найти слово, которое точно присутствует в документе.
6. Проверить имя файла, страницу, фрагмент, подсветку и релевантность.
7. Проверить переход на вторую страницу выдачи.
8. Проверить пустой запрос, отсутствие результатов и загрузку неверного формата.

Если backend соблюдает описанные контракты, дополнительных изменений в интерфейсе для первой интеграции не потребуется.
