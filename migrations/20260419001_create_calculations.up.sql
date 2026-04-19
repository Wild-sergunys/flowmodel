SET NAMES utf8mb4;

CREATE TABLE calculations (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  material_id INT NOT NULL,
  input_json JSON NOT NULL,
  result_json JSON NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (material_id) REFERENCES materials(id) ON DELETE CASCADE
);
