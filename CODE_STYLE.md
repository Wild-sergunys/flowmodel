# Code Style Guide — flowmodel

## Go

### Именование
| Тип        | Правило               | Пример                         |
| ---------- | --------------------- | ------------------------------ |
| Пакет      | lowercase, одно слово | rheology, thermal              |
| Файл       | snake_case            | file.go, big_file.go           |
| Структура  | PascalCase            | CalculationInput, MaterialRepo |
| Интерфейс  | PascalCase + er       | Solver, Repository             |
| Функция    | PascalCase (публ)     | Calculate(), validateInput()   |
| Переменная | camelCase             | materialID, stepCount          |
| Константа  | PascalCase            | MaxSteps, DefaultTemp          |

### Обработка ошибок

```go
// Всегда проверяем ошибки
result, err := repo.GetMaterial(id)
if err != nil {
    return nil, fmt.Errorf("failed to get material: %w", err)
}

// Не игнорируем
result, _ := repo.GetMaterial(id) // так нельзя
```
### Комментарии

```go
// CalculateViscosity вычисляет вязкость по формуле Андраде.
// T — температура в Кельвинах.
// gammaDot — скорость деформации сдвига.
func CalculateViscosity(mu0, T, gammaDot float64) float64 {
    // ...
}
```
### Форматирование
- TS - 2 пробела
- Максимальная длина строки — 120 символов
- Импорты: стандартная библиотека -> внешние -> внутренние
```go
import (
	"fmt"
	"math"
	
	"github.com/go-sql-driver/mysql"
    
    "flowmodel/internal/model"
)
```

---

## SQL (Миграции)

### Именование
Тип          | Правило                    | Пример
-------------|----------------------------|----------------------------
Файл миграции| NNN_description.sql        | 001_create_users.sql
Таблица      | snake_case, множ. число    | materials, parameters
Поле         | snake_case, ед. число      | created_at, material_id
Первичный ключ| id                        | id INT AUTO_INCREMENT PRIMARY KEY
Внешний ключ | {table}_id                 | material_id

### Пример миграции
``` sql
-- 001_create_materials.sql
CREATE TABLE materials (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

---
## Git

### Ветки
- main — только рабочий, проверенный код
- feature/имя-фичи — разработка новой функциональности
- fix/что-чиним — исправление багов

### Коммиты
<префикс>: <описание на русском>

feat: реализован расчёт вязкости по Андраде
fix: исправлено деление на ноль в thermal/balance.go
chore: обновлён .gitignore
docs: добавлен CODE_STYLE.md
refactor: вынесен MaterialRepository в интерфейс
test: добавлены тесты для solver.go


---
## Комментарии в коде

### На русском
``` go
// CalculateTemperatureProfile вычисляет распределение температуры по длине канала.
func CalculateTemperatureProfile(input model.CalculationInput) ([]model.Point, error) {
    dx := input.L / float64(input.Steps)
    // ...
}
```
---
## Соглашения по проекту

Договорённость         | Значение
-----------------------|----------------
Язык комментариев      | Русский
Язык логов             | Русский
Формат даты в БД       | TIMESTAMP
Порт по умолчанию      | 8080
Точка входа            | cmd/server/main.go
