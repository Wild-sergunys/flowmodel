-- 20260417_003_create_parameters.up.sql

CREATE TABLE parameters (
  id INT AUTO_INCREMENT PRIMARY KEY,
  code VARCHAR(50) NOT NULL UNIQUE COMMENT 'Технический код',
  name VARCHAR(255) NOT NULL COMMENT 'Человеческое название',
  unit VARCHAR(50) COMMENT 'Единица измерения',
  data_type ENUM('float', 'int', 'string') NOT NULL DEFAULT 'float',
  category ENUM('material_property', 'empirical_coefficient', 'process_parameter') NOT NULL,
  description TEXT COMMENT 'Описание параметра',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);