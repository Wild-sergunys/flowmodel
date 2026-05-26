-- MySQL dump 10.13  Distrib 8.0.45, for Linux (x86_64)
--
-- Host: localhost    Database: flowmodel
-- ------------------------------------------------------
-- Server version	8.0.45

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `calculations`
--

DROP TABLE IF EXISTS `calculations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `calculations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `material_id` int NOT NULL,
  `input_json` json NOT NULL,
  `result_json` json NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `material_id` (`material_id`),
  CONSTRAINT `calculations_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `calculations_ibfk_2` FOREIGN KEY (`material_id`) REFERENCES `materials` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `calculations`
--

LOCK TABLES `calculations` WRITE;
/*!40000 ALTER TABLE `calculations` DISABLE KEYS */;
/*!40000 ALTER TABLE `calculations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `material_parameters`
--

DROP TABLE IF EXISTS `material_parameters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `material_parameters` (
  `material_id` int NOT NULL,
  `parameter_id` int NOT NULL,
  `value_float` double DEFAULT NULL COMMENT 'Числовое значение',
  `value_string` text COLLATE utf8mb4_unicode_ci COMMENT 'Строковое значение',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`material_id`,`parameter_id`),
  KEY `parameter_id` (`parameter_id`),
  CONSTRAINT `material_parameters_ibfk_1` FOREIGN KEY (`material_id`) REFERENCES `materials` (`id`) ON DELETE CASCADE,
  CONSTRAINT `material_parameters_ibfk_2` FOREIGN KEY (`parameter_id`) REFERENCES `parameters` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `material_parameters`
--

LOCK TABLES `material_parameters` WRITE;
/*!40000 ALTER TABLE `material_parameters` DISABLE KEYS */;
INSERT INTO `material_parameters` VALUES (1,1,1380,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,2,2500,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,3,145,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,4,12000,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,5,147000,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,6,180,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,7,0.28,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(1,8,400,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,1,950,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,2,2300,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,3,135,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,4,15000,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,5,140000,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,6,190,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,7,0.35,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,8,350,NULL,'2026-05-26 10:30:58','2026-05-26 10:30:58'),(3,1,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,2,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,3,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,4,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,5,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,6,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,7,0.5,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29'),(3,8,123,NULL,'2026-05-26 10:33:29','2026-05-26 10:33:29');
/*!40000 ALTER TABLE `material_parameters` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `materials`
--

DROP TABLE IF EXISTS `materials`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `materials` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Название материала',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'Описание материала',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `materials`
--

LOCK TABLES `materials` WRITE;
/*!40000 ALTER TABLE `materials` DISABLE KEYS */;
INSERT INTO `materials` VALUES (1,'ПВХ','Поливинилхлорид','2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,'ПЭВП','Полиэтилен высокой плотности','2026-05-26 10:30:58','2026-05-26 10:30:58'),(3,'test','123','2026-05-26 10:32:42','2026-05-26 10:32:42');
/*!40000 ALTER TABLE `materials` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `parameters`
--

DROP TABLE IF EXISTS `parameters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `parameters` (
  `id` int NOT NULL AUTO_INCREMENT,
  `code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Технический код',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Человеческое название',
  `unit` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Единица измерения',
  `data_type` enum('float','int','string') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'float',
  `category` enum('material_property','empirical_coefficient','process_parameter') COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'Описание параметра',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `parameters`
--

LOCK TABLES `parameters` WRITE;
/*!40000 ALTER TABLE `parameters` DISABLE KEYS */;
INSERT INTO `parameters` VALUES (1,'density','Плотность','кг/м³','float','material_property','Масса единицы объёма. Должна быть > 0','2026-05-26 10:30:58'),(2,'heat_capacity','Удельная теплоёмкость','Дж/(кг·°С)','float','material_property','Количество теплоты для нагрева 1 кг на 1°С. Должна быть > 0','2026-05-26 10:30:58'),(3,'melting_temp','Температура плавления','°С','float','material_property','Температура начала плавления. Должна быть > 0','2026-05-26 10:30:58'),(4,'mu0','Коэффициент консистенции μ0','Па·с^n','float','empirical_coefficient','Вязкость при температуре приведения. Должна быть > 0','2026-05-26 10:30:58'),(5,'Ea','Энергия активации вязкого течения','Дж/моль','float','empirical_coefficient','Энергия активации. Должна быть > 0','2026-05-26 10:30:58'),(6,'Tr','Температура приведения','°С','float','empirical_coefficient','Температура приведения. Должна быть > 0','2026-05-26 10:30:58'),(7,'n','Индекс течения','','float','empirical_coefficient','Индекс течения (0 < n < 1 для псевдопластиков)','2026-05-26 10:30:58'),(8,'alpha_u','Коэффициент теплоотдачи','Вт/(м²·°С)','float','process_parameter','Коэффициент теплоотдачи от крышки. Должен быть ≥ 0','2026-05-26 10:30:58');
/*!40000 ALTER TABLE `parameters` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `login` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `role` enum('researcher','admin') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'researcher',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `login` (`login`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'admin','$2a$10$dbGDmlZf4wi74l3FfTJyU./jGCVXliu59pyYmXbmCrTXXQuPz9meu','admin','2026-05-26 10:30:58','2026-05-26 10:30:58'),(2,'user','$2a$10$H895pcOZdUFpLPel/K7jxO5WZCcNZTDD2dmfzncTzU6VihC6dc6g6','researcher','2026-05-26 10:31:45','2026-05-26 10:31:45');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-05-26 10:35:03
