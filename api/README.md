# Flowmodel API

API для моделирования неизотермического течения аномально-вязких материалов.

## Быстрый старт

Для запуска потребуется установленный Docker и Docker Compose.

### 1. Клонировать репозиторий

    git clone https://github.com/Wild-sergunys/flowmodel.git
    cd flowmodel

### 2. Создать .env

    cp .env.example .env

При необходимости отредактируйте значения в .env.

### 3. Запустить проект

    docker compose up -d


## Аутентификация

Используются JWT токены. После успешного входа сервер возвращает токен, который клиент должен передавать в заголовке Authorization: Bearer <token> при каждом запросе к защищённым эндпоинтам.

Токен действителен 24 часа.

Учётные данные по умолчанию:
- Логин: admin
- Пароль: admin123

Ответ при успешном входе:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "role": "admin"
}

## Эндпоинты

### Авторизация (публичные)

| Метод | URL | Описание |
|-------|-----|----------|
| POST | /api/auth/login | Вход в систему |
| POST | /api/auth/logout | Выход |
| GET | /api/auth/me | Информация о текущем пользователе |

### Валидация (публичный)

| Метод | URL | Описание |
|-------|-----|----------|
| POST | /api/validate | Валидация входных параметров |

### Расчёт (researcher, admin)

| Метод | URL | Описание |
|-------|-----|----------|
| POST | /api/calculate | Расчёт параметров течения |

Пример запроса:
{
  "w": 0.25,
  "h": 0.01,
  "l": 9.5,
  "vu": 1.5,
  "tu": 150,
  "material_id": 1,
  "t0": 145,
  "steps": 100
}

Пример ответа:
{
  "id": 42,
  "productivity": 0.001875,
  "temperature": 152.3,
  "viscosity": 8450.2,
  "profile": [
    {"x": 0, "temperature": 145, "viscosity": 12000},
    {"x": 0.095, "temperature": 146.1, "viscosity": 11800}
  ],
  "metrics": {
    "calc_time_ms": 15,
    "memory_used_bytes": 2048
  }
}

### Справочник (researcher, admin)

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/materials | Список материалов |

Пример ответа:
[{"id": 1, "name": "ПВХ"}]

### История расчётов (researcher, admin)

| Метод | URL | Описание |
|-------|-----|----------|
| GET | /api/results | История расчётов текущего пользователя |
| GET | /api/results/{id} | Получить расчёт по ID |
| GET | /api/results/{id}/report | Получить отчёт о расчёте |
| GET | /api/results/{id}/download | Скачать отчёт в Excel (.xlsx) |

### Администрирование (только admin)

| Метод | URL | Описание |
|-------|-----|----------|
| Материалы |||
| GET | /api/admin/materials | Список материалов |
| POST | /api/admin/materials | Создать материал |
| PUT | /api/admin/materials/{id} | Обновить материал |
| DELETE | /api/admin/materials/{id} | Удалить материал |
| GET | /api/admin/materials/{id}/parameters | Получить значения параметров |
| PUT | /api/admin/materials/{id}/parameters | Обновить значения параметров |
| Параметры |||
| GET | /api/admin/parameters | Список параметров |
| POST | /api/admin/parameters | Создать параметр |
| PUT | /api/admin/parameters/{id} | Обновить параметр |
| DELETE | /api/admin/parameters/{id} | Удалить параметр |
| Пользователи |||
| GET | /api/admin/users | Список пользователей |
| POST | /api/admin/users | Создать пользователя |
| PUT | /api/admin/users/{id} | Обновить пользователя |
| DELETE | /api/admin/users/{id} | Удалить пользователя |

## Коды ответов

| Код | Описание |
|-----|----------|
| 200 | Успех |
| 201 | Создано |
| 400 | Ошибка валидации |
| 401 | Не авторизован |
| 403 | Нет прав |
| 404 | Не найдено |
| 429 | Слишком много попыток входа |
| 500 | Ошибка сервера |

## OpenAPI спецификация

Полная спецификация в формате OpenAPI 3.0 доступна в файле openapi.yaml.