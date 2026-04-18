-- 20260417_001_create_users.up.sql

SET NAMES utf8mb4;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    login VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('researcher', 'admin') NOT NULL DEFAULT 'researcher',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Администратор по умолчанию (пароль: admin123)
INSERT INTO users (login, password_hash, role) VALUES 
('admin', '$2a$10$dbGDmlZf4wi74l3FfTJyU./jGCVXliu59pyYmXbmCrTXXQuPz9meu', 'admin');