package service

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// GenerateExcel создаёт отчёт в формате Excel
func GenerateExcel(calcData map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Лист 1 — Результаты
	f.SetSheetName("Sheet1", "Результаты моделирования")

	// Заголовок
	f.SetCellValue("Результаты моделирования", "A1", "FLOWMODEL - Отчёт о моделировании течения")
	f.SetCellValue("Результаты моделирования", "A2", fmt.Sprintf("Дата и время: %s", time.Now().Format("02.01.2006 15:04:05")))

	// Объединяем ячейки для заголовка
	f.MergeCell("Результаты моделирования", "A1", "F1")
	f.MergeCell("Результаты моделирования", "A2", "F2")

	// Стили
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#1a1a1a"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "#ffffff", Size: 10},
	})

	boldStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10},
	})

	// Раздел 1: Входные данные
	row := 4
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "ВХОДНЫЕ ПАРАМЕТРЫ")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), boldStyle)
	row += 2

	input := calcData["input"].(map[string]interface{})

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Параметр")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	row++

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
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), boldStyle)
	row += 2

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Параметр")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Обозначение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++

	materialName := "ПВХ"
	if mn, ok := calcData["material_name"].(string); ok {
		materialName = mn
	}

	materialRows := [][]interface{}{
		{"Название материала", "-", materialName, "-"},
		{"Плотность", "ρ", 1380, "кг/м³"},
		{"Удельная теплоёмкость", "c", 2500, "Дж/(кг·°С)"},
		{"Температура плавления", "T0", 145, "°С"},
		{"Коэффициент консистенции", "μ0", 12000, "Па·с^n"},
		{"Энергия активации", "Ea", 147000, "Дж/моль"},
		{"Температура приведения", "Tr", 180, "°С"},
		{"Индекс течения", "n", 0.28, "-"},
		{"Коэффициент теплоотдачи", "αu", 400, "Вт/(м²·°С)"},
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
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), boldStyle)
	row += 2

	result := calcData["result"].(map[string]interface{})

	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Показатель")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), "Обозначение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("C%d", row), "Значение")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("D%d", row), "Ед. изм.")
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), headerStyle)
	row++

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
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), boldStyle)
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
	f.SetCellStyle("Результаты моделирования", fmt.Sprintf("A%d", row), fmt.Sprintf("F%d", row), boldStyle)
	row += 2

	metrics, _ := result["metrics"].(map[string]interface{})
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Время расчёта")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%v мс", metrics["calc_time_ms"]))
	row++

	memoryBytes, _ := metrics["memory_used_bytes"].(float64)
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Использовано памяти")
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("B%d", row), fmt.Sprintf("%.0f КБ", memoryBytes/1024))

	// Раздел 6: Подпись
	row += 2
	f.SetCellValue("Результаты моделирования", fmt.Sprintf("A%d", row), "Отчёт сформирован автоматически программным комплексом FlowModel")

	// Автоширина колонок
	for col := 'A'; col <= 'F'; col++ {
		f.SetColWidth("Результаты моделирования", string(col), string(col), 25)
	}

	// Сохраняем в буфер
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Excel: %w", err)
	}

	return buf.Bytes(), nil
}
