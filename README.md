FLOWMODEL — РАСЧЁТ ПАРАМЕТРОВ КАНАЛА

Веб-приложение для моделирования неизотермического течения
аномально-вязких материалов в канале с подвижной крышкой.


СТЕК

Backend:    Go 1.25, стандартный net/http, JWT-аутентификация
База:       MySQL 8.0
Фронтенд:   Встроенные шаблоны, минимальный JS
Деплой:     Docker, Docker Compose


ВОЗМОЖНОСТИ

- Моделирование течения аномально-вязких материалов в канале
- Расчёт производительности, температуры и вязкости по длине канала
- Учёт параметров материала (энергия активации, индекс течения, плотность и др.)
- Валидация входных параметров
- Сохранение истории расчётов
- Экспорт результатов в Excel и JSON
- Административная панель управления материалами, параметрами и пользователями
- Аутентификация и разделение ролей (admin, researcher)


АРХИТЕКТУРА

.
├── api
│   ├── FORMATS.md
│   ├── openapi.yaml
│   ├── README.md
│   ├── REPORT.md
│   └── VISUALIZATION.md
├── cmd
│   └── server
│       └── main.go
├── CODE_STYLE.md
├── compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── config.go
│   ├── database
│   │   ├── connection.go
│   │   └── migrate.go
│   ├── handler
│   │   ├── admin.go
│   │   ├── auth.go
│   │   ├── calculation.go
│   │   ├── material.go
│   │   ├── material_parameter.go
│   │   ├── parameter.go
│   │   ├── response.go
│   │   ├── results.go
│   │   └── user.go
│   ├── middleware
│   │   ├── auth.go
│   │   ├── rate_limit.go
│   │   └── rate_limit_test.go
│   ├── model
│   │   ├── calculation.go
│   │   ├── error.go
│   │   ├── material.go
│   │   ├── parameter.go
│   │   └── user.go
│   ├── repository
│   │   ├── calculation.go
│   │   ├── material.go
│   │   ├── material_parameter.go
│   │   ├── parameter.go
│   │   └── user.go
│   ├── service
│   │   ├── auth.go
│   │   ├── auth_test.go
│   │   ├── calculation.go
│   │   └── report.go
│   └── validator
│       ├── calculation.go
│       └── calculation_test.go
├── LICENSE
├── migrations
│   ├── 20260417001_create_users.down.sql
│   ├── 20260417001_create_users.up.sql
│   ├── 20260417002_create_materials.down.sql
│   ├── 20260417002_create_materials.up.sql
│   ├── 20260417003_create_parameters.down.sql
│   ├── 20260417003_create_parameters.up.sql
│   ├── 20260417004_create_material_parameters.down.sql
│   ├── 20260417004_create_material_parameters.up.sql
│   ├── 20260419001_create_calculations.down.sql
│   └── 20260419001_create_calculations.up.sql
├── README.md
└── web
    ├── frontend.go
    ├── frontend_test.go
    ├── static
    │   ├── css
    │   │   └── app.css
    │   └── js
    │       ├── api.js
    │       ├── app.js
    │       └── pages
    │           ├── admin.js
    │           ├── cabinet.js
    │           ├── home.js
    │           └── login.js
    └── templates
        ├── layout.html
        ├── pages
        │   ├── admin.html
        │   ├── cabinet.html
        │   ├── home.html
        │   └── login.html
        └── partials
            └── header.html



ФИЗИЧЕСКАЯ МОДЕЛЬ

Приложение моделирует течение расплава полимера в канале экструдера
с подвижной верхней крышкой:

  - Геометрия канала: ширина W, глубина H, длина L
  - Скорость крышки Vu, температура крышки Tu
  - Материал описывается параметрами: mu0, Ea, Tr, n, alpha_u,
    density, heat_capacity, melting_temp
  - Учитывается диссипативный нагрев от вязкого трения
  - Расчёт профиля температур и вязкости по длине канала
  - Результат: производительность (кг/ч), температура и вязкость


ЗАЩИТА

- JWT-аутентификация с разделением ролей
- Rate limiter на попытки входа (защита от брутфорса)
- Graceful shutdown сервера

