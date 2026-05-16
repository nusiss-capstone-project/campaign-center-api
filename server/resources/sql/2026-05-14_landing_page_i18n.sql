-- Landing page i18n: default_lang + translation table
-- For existing databases that still have column `language` on campaign_landing_pages.
-- Skip if you already applied a schema that includes `default_lang` and campaign_landing_page_translations.
-- Target: MySQL 8.x

USE `campaign_center`;

-- Rename language -> default_lang (align with translation_design.pdf)
ALTER TABLE `campaign_landing_pages`
  CHANGE COLUMN `language` `default_lang` VARCHAR(16) DEFAULT 'en';

CREATE TABLE IF NOT EXISTS `campaign_landing_page_translations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `landing_page_id` BIGINT NOT NULL,
  `lang` VARCHAR(16) NOT NULL,
  `title` VARCHAR(255) DEFAULT NULL,
  `description` TEXT,
  `terms` TEXT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `created_by` VARCHAR(255) NOT NULL DEFAULT '',
  `updated_by` VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_landing_page_lang` (`landing_page_id`, `lang`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
