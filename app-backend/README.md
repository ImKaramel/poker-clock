# API Documentation

Документация REST API для приложения


## Авторизация

### Telegram Auth

**POST** `/api/auth/telegram/`

Аутентификация через данные Telegram (WebApp или Mini App).

**Request:**
```json
{
  "telegram_data": { ... }
}
```

Response:

```json
{
  "token": "jwt_token",
  "user": { ... },
  "is_new": true
}
```

Валидация initData
POST /api/auth/telegram/validate/

Валидация initData от Telegram WebApp.

Request:

```json
{
  "initData": "string"
}
Response:

json
{
  "token": "jwt_token",
  "user": {
    "username": "...",
    "first_name": "...",
    "last_name": "...",
    "telegram_id": "...",
    "id": "..."
  }
}
```

Общие заголовки
Для всех защищённых эндпоинтов обязателен заголовок:

```Authorization: Bearer <jwt_token>```

Ошибки
Все ошибки возвращаются в едином формате:

```json
{
  "error": "Описание ошибки"
}
```
 
### Пользователи (только для Admin)
```
Метод	Эндпоинт	Описание
GET	/api/users	Список всех пользователей
POST	/api/users	Создать пользователя
GET	/api/users/:user_id	Получить пользователя
PATCH	/api/users/:user_id	Обновить пользователя
DELETE	/api/users/:user_id	Удалить пользователя
POST	/api/users/:user_id/ban	Заблокировать пользователя
POST	/api/users/:user_id/unban	Разблокировать пользователя
POST	/api/users/:user_id/add_points	Добавить баллы
```
Пример добавления баллов:
Request:

```json
{
  "points": 100
}
``` 
### Игры (Games)
```
Метод	Эндпоинт	Описание	Права
GET	/api/games	Список игр	Все
GET	/api/games/:id	Получить игру	Admin
POST	/api/games	Создать игру	Admin
PATCH	/api/games/:id	Обновить игру	Admin
DELETE	/api/games/:id	Удалить игру	Admin
```
Создание игры (Request):
```json
{
  "date": "2026-04-05",
  "time": "19:00:00",
  "description": "Еженедельный турнир",
  "buyin": 100,
  "location": "Москва"
}
``` 
### Участники игры (Admin)
```
Метод	Эндпоинт	Описание
GET	/api/games/:id/participants_admin	Список участников (админ)
POST	/api/games/:id/add_participant_admin	Добавить участника
POST	/api/games/:id/remove_participant_admin	Удалить участника
POST	/api/games/:id/update_participant_admin	Обновить данные участника
POST	/api/games/:id/complete	Завершить турнир
```
Добавление участника:
```json
{
  "user_id": "string",
  "entries": 1,
  "rebuys": 0,
  "addons": 0
}
```
### Участники (Participants)
```
Метод	Эндпоинт	Описание
GET	/api/participants	Список своих участников (или всех для admin)
POST	/api/participants/register	Зарегистрироваться на игру
DELETE	/api/participants/unregister	Отменить регистрацию
GET	/api/participants/:id	Получить участника
POST	/api/participants	Создать запись
PATCH	/api/participants/:id	Обновить запись
DELETE	/api/participants/:id	Удалить запись
```

### История турниров (Admin only)

```
Метод	Эндпоинт	Описание
GET	/api/tournament-history	Список истории
POST	/api/tournament-history	Создать запись
GET	/api/tournament-history/:id	Детали турнира
PATCH	/api/tournament-history/:id	Обновить
DELETE	/api/tournament-history/:id	Удалить
GET	/api/tournament-history/:id/participants	Участники турнира
```
### Тикеты поддержки (Support Tickets)
```Метод	Эндпоинт	Описание
GET	/api/support-tickets	Список тикетов
POST	/api/support-tickets	Создать тикет
GET	/api/support-tickets/:id	Получить тикет
PATCH	/api/support-tickets/:id	Обновить тикет
DELETE	/api/support-tickets/:id	Удалить тикет
```
### Доп эндпоинты
```Метод	Эндпоинт	Описание
GET	/api/rating	Рейтинг пользователей
GET	/api/profile	Получить свой профиль
PATCH	/api/profile	Обновить свой профиль
GET	/api/admin/dashboard	Админская панель
```