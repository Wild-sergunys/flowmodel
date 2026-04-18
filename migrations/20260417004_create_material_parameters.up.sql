-- 20260417_004_create_material_parameters.up.sql

SET NAMES utf8mb4;

CREATE TABLE material_parameters (
  material_id INT NOT NULL,
  parameter_id INT NOT NULL,
  value_float DOUBLE COMMENT 'Числовое значение',
  value_string TEXT COMMENT 'Строковое значение',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (material_id, parameter_id),
  FOREIGN KEY (material_id) REFERENCES materials(id) ON DELETE CASCADE,
  FOREIGN KEY (parameter_id) REFERENCES parameters(id) ON DELETE CASCADE
);

-- НАЧАЛЬНЫЕ ДАННЫЕ (СИДЫ)

-- Материал ПВХ
INSERT INTO materials (id, name, description) VALUES 
(1, 'ПВХ', 'Поливинилхлорид');

-- Параметры
INSERT INTO parameters (id, code, name, unit, data_type, category) VALUES
(1, 'density', 'Плотность', 'кг/м³', 'float', 'material_property'),
(2, 'heat_capacity', 'Удельная теплоёмкость', 'Дж/(кг·°С)', 'float', 'material_property'),
(3, 'melting_temp', 'Температура плавления', '°С', 'float', 'material_property'),
(4, 'mu0', 'Коэффициент консистенции μ0', 'Па·с^n', 'float', 'empirical_coefficient'),
(5, 'Ea', 'Энергия активации вязкого течения', 'Дж/моль', 'float', 'empirical_coefficient'),
(6, 'Tr', 'Температура приведения', '°С', 'float', 'empirical_coefficient'),
(7, 'n', 'Индекс течения', '', 'float', 'empirical_coefficient'),
(8, 'alpha_u', 'Коэффициент теплоотдачи от крышки', 'Вт/(м²·°С)', 'float', 'process_parameter');

-- Значения для ПВХ
INSERT INTO material_parameters (material_id, parameter_id, value_float) VALUES
(1, 1, 1380),
(1, 2, 2500),
(1, 3, 145),
(1, 4, 12000),
(1, 5, 147000),
(1, 6, 180),
(1, 7, 0.28),
(1, 8, 400);