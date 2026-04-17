-- 20260417_004_create_material_parameters.up.sql

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