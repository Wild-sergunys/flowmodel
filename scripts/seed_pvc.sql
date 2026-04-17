-- Начальные данные для тестирования

-- Материал ПВХ
INSERT INTO materials (id, name, description) VALUES 
(1, 'ПВХ', 'Поливинилхлорид')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Параметры
INSERT INTO parameters (id, code, name, unit, data_type, category) VALUES
(1, 'density', 'Плотность', 'кг/м³', 'float', 'material_property'),
(2, 'heat_capacity', 'Удельная теплоёмкость', 'Дж/(кг·°С)', 'float', 'material_property'),
(3, 'melting_temp', 'Температура плавления', '°С', 'float', 'material_property'),
(4, 'mu0', 'Коэффициент консистенции μ0', 'Па·с^n', 'float', 'empirical_coefficient'),
(5, 'Ea', 'Энергия активации вязкого течения', 'Дж/моль', 'float', 'empirical_coefficient'),
(6, 'Tr', 'Температура приведения', '°С', 'float', 'empirical_coefficient'),
(7, 'n', 'Индекс течения', '', 'float', 'empirical_coefficient'),
(8, 'alpha_u', 'Коэффициент теплоотдачи от крышки', 'Вт/(м²·°С)', 'float', 'process_parameter')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Значения для ПВХ
INSERT INTO material_parameters (material_id, parameter_id, value_float) VALUES
(1, 1, 1380),
(1, 2, 2500),
(1, 3, 145),
(1, 4, 12000),
(1, 5, 147000),
(1, 6, 180),
(1, 7, 0.28),
(1, 8, 400)
ON DUPLICATE KEY UPDATE value_float = VALUES(value_float);