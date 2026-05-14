package service

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateExcel(calcData map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Лист 1 - Результаты
	f.SetSheetName("Sheet1", "Результаты моделирования")

	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#1a1a1a"}, Pattern: 1},
		Font:   &excelize.Font{Bold: true, Color: "#ffffff", Size: 10},
		Border: []excelize.Border{{Type: "bottom", Color: "#000000", Style: 2}},
	})

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true, Size: 10},
		Border: []excelize.Border{{Type: "bottom", Color: "#000000", Style: 1}},
	})

	// Заголовок
	f.SetCellValue("Результаты моделирования", "A1", "FLOWMODEL - Отчёт о моделировании течения")
	f.SetCellValue("Результаты моделирования", "A2", fmt.Sprintf("Дата и время: %s", time.Now().Format("02.01.2006 15:04:05")))

	// Объединяем ячейки для заголовка
	f.MergeCell("Результаты моделирования", "A1", "D1")
	f.MergeCell("Результаты моделирования", "A2", "D2")

	row := 4

	// Раздел 1: Входные данные
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "ВХОДНЫЕ ПАРАМЕТРЫ")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), boldStyle)
	row += 2

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Параметр")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	row++

	// Получаем входные данные
	input := calcData["input"].(map[string]interface{})

	inputRows := [][]interface{}{
		{"Ширина канала, W", input["w"], "м"},
		{"Глубина канала, H", input["h"], "м"},
		{"Длина канала, L", input["l"], "м"},
		{"Скорость крышки, Vu", input["vu"], "м/с"},
		{"Температура крышки, Tu", input["tu"], "°C"},
		{"Количество шагов", input["steps"], "-"},
	}

	for _, r := range inputRows {
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), r[0])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), r[1])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), r[2])
		row++
	}

	// Раздел 2: Параметры материала
	row++
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "ПАРАМЕТРЫ МАТЕРИАЛА")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), boldStyle)
	row += 2

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Параметр")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Обозначение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++

	// Получаем имя материала
	materialName := "неизвестный"
	if mn, ok := calcData["material_name"].(string); ok {
		materialName = mn
	}

	// Пытаемся извлечь параметры материала из входных данных или из params
	var materialParams map[string]interface{}
	if params, ok := calcData["material_params"].(map[string]interface{}); ok {
		materialParams = params
	}

	// Заполняем параметры материала
	materialRows := [][]interface{}{
		{"Название материала", "-", materialName, "-"},
	}

	// Конфигурация параметров с их единицами измерения
	paramConfigs := []struct {
		label      string
		symbol     string
		key        string
		unit       string
		defaultVal float64
	}{
		{"Плотность", "ρ", "density", "кг/м³", 1380},
		{"Удельная теплоёмкость", "c", "heat_capacity", "Дж/(кг·°С)", 2500},
		{"Температура плавления", "T0", "melting_temp", "°С", 145},
		{"Коэффициент консистенции", "μ0", "mu0", "Па·сⁿ", 12000},
		{"Энергия активации", "Ea", "Ea", "Дж/моль", 147000},
		{"Температура приведения", "Tr", "Tr", "°С", 180},
		{"Индекс течения", "n", "n", "-", 0.28},
		{"Коэффициент теплоотдачи", "αu", "alpha_u", "Вт/(м²·°С)", 400},
	}

	for _, cfg := range paramConfigs {
		var value float64
		if materialParams != nil {
			if v, ok := materialParams[cfg.key]; ok {
				switch vv := v.(type) {
				case float64:
					value = vv
				case int:
					value = float64(vv)
				}
			}
		}
		if value == 0 {
			value = cfg.defaultVal
		}
		materialRows = append(materialRows, []interface{}{cfg.label, cfg.symbol, value, cfg.unit})
	}

	for _, r := range materialRows {
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), r[0])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), r[1])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), r[2])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), r[3])
		row++
	}

	// Раздел 3: Результаты расчёта
	row++
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "РЕЗУЛЬТАТЫ РАСЧЁТА")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), boldStyle)
	row += 2

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Показатель")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Обозначение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++

	result := calcData["result"].(map[string]interface{})

	resultRows := [][]interface{}{
		{"Производительность", "Q", fmt.Sprintf("%.2f", result["productivity"]), "кг/ч"},
		{"Температура продукта", "Tp", fmt.Sprintf("%.1f", result["temperature"]), "°C"},
		{"Вязкость продукта", "ηp", fmt.Sprintf("%.1f", result["viscosity"]), "Па·с"},
	}

	for _, r := range resultRows {
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), r[0])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), r[1])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), r[2])
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), r[3])
		row++
	}

	// Раздел 4: Таблица распределений
	row++
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "ТАБЛИЦА РАСПРЕДЕЛЕНИЙ ПО ДЛИНЕ КАНАЛА")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), boldStyle)
	row += 2

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Координата X, м")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Температура T, °C")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Вязкость η, Па·с")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	row++

	profile, ok := result["profile"].([]interface{})
	if ok {
		for _, p := range profile {
			point := p.(map[string]interface{})
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("%.4f", point["x"]))
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%.1f", point["temperature"]))
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f", point["viscosity"]))
			row++
		}
	}

	// Раздел 5: Метрики
	row++
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "МЕТРИКИ ПРОИЗВОДИТЕЛЬНОСТИ")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), boldStyle)
	row += 2

	metrics, _ := result["metrics"].(map[string]interface{})
	if metrics != nil {
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Время расчёта")
		f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%v мс", metrics["calc_time_ms"]))
		row++

		memoryBytes, _ := metrics["memory_used_bytes"].(float64)
		memoryKB := memoryBytes / 1024
		if memoryKB < 1024 {
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Использовано памяти")
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%.0f КБ", memoryKB))
		} else {
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Использовано памяти")
			f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f МБ", memoryKB/1024))
		}
	}

	// Раздел 6: Подпись
	row += 2
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Отчёт сформирован автоматически программным комплексом FlowModel")

	// Автоширина колонок
	f.SetColWidth("Результаты моделирования", "A", "A", 28)
	f.SetColWidth("Результаты моделирования", "B", "B", 15)
	f.SetColWidth("Результаты моделирования", "C", "C", 15)
	f.SetColWidth("Результаты моделирования", "D", "D", 18)

	// Сохраняем в буфер
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Excel: %w", err)
	}

	return buf.Bytes(), nil
}
