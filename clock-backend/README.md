#  Poker Clock Backend API

---

#  Base URL

```
http://localhost:8081
```

---

#  Health Check

## GET `/health`

Проверка работоспособности сервера.

### Response

```json
"ok"
```

---

# Authentication

Все защищённые эндпоинты требуют заголовок:

Authorization: Bearer <ADMIN_PASSWORD>

Где `<ADMIN_PASSWORD>` — значение переменной окружения:

ADMIN_PASSWORD=your_secret_password

### Response

* `200 OK` — успешная авторизация
* `401 Unauthorized` — неверный пароль

---

#  Tournaments

## POST `/tournaments/`

Создать турнир (только для администратора)

### Request

```json
{
  "name": "Example"
}
```

### Response

```json
{
  "id": "string",
  "name": "string",
  "levels": []
}
```

---

## GET `/tournaments/`

Получить список турниров

### Response

```json
[
  {
    "id": "string",
    "name": "string",
    "levels": []
  }
]
```

---

## GET `/tournaments/{id}`

Получить турнир по ID

### Response

```json
{
  "id": "string",
  "name": "string",
  "levels": [
    {
      "small_blind": 10,
      "big_blind": 20,
      "duration_minutes": 15
    }
  ]
}
```

---

#  Levels

## POST `/tournaments/{id}/levels`

Добавить уровень в турнир

### Request

```json
{
  "small_blind": 10,
  "big_blind": 20,
  "duration_minutes": 15
}
```

### Response

```json
{
  "id": "string",
  "name": "string",
  "levels": [...]
}
```

---

## GET `/tournaments/{id}/levels`

Получить список уровней турнира

### Response

```json
[
  {
    "small_blind": 10,
    "big_blind": 20,
    "duration_minutes": 15
  }
]
```

---

#  Timer

## POST `/tournaments/{id}/start`

Запустить таймер турнира

### Response

* `204 No Content`

---

## POST `/tournaments/{id}/pause`

Поставить таймер на паузу

### Response

* `204 No Content`

---

## POST `/tournaments/{id}/resume`

Возобновить таймер

### Response

* `204 No Content`

---

## POST `/tournaments/{id}/next`

Переключить на следующий уровень

### Response

* `204 No Content`

---

#  WebSocket (Timer Updates)

## GET `/tournaments/{id}/timer/ws`

Подписка на обновления таймера в реальном времени.

### Initial message

```json
{
  "level": 1,
  "small_blind": 10,
  "big_blind": 20,
  "remaining_seconds": 900
}
```

### Update message

```json
{
  "level": 2,
  "small_blind": 20,
  "big_blind": 40,
  "remaining_seconds": 850
}
```

---

### Error Responses

### Общий формат

```json
{
  "error": "message"
}
```


# TODO 
ручка добавлена --- на стороне часов только  --- нужна доработка с app-backend
## POST `/tournaments/{id}/stats`

Обновить статистику турнира (игроки, фишки)

### Request

```json
{
  "players_count": 10,
  "total_chips": 15000
}
```
---

#  Пример использования

### 1. Логин

```
POST /auth/login
```
(впринципе не нужен -- вход просто по паролю )

### 2. Создать турнир

```
POST /tournaments/
```

### 3. Добавить уровни

```
POST /tournaments/{id}/levels
```

### 4. Запустить таймер

```
POST /tournaments/{id}/start
```

### 5. Подключиться к WebSocket

```
ws://localhost:8081/tournaments/{id}/timer/ws
```

---

### Example Request
```
curl -X POST http://localhost:8081/tournaments/ \
-H "Authorization: Bearer secretFromEnv" \
-H "Content-Type: application/json" \
-d '{"name": "Sunday Poker"}'
```
#  Notes

* Таймер работает **в runtime (in-memory)** и обновляется через WebSocket
* Состояние таймера частично хранится в БД

---


