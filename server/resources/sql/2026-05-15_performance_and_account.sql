-- Campaign performance, user account & unified transactions
-- Target: MySQL 8.x, utf8mb4 (additive; does not alter existing tables)

SET NAMES utf8mb4;

USE `campaign_center`;

CREATE TABLE IF NOT EXISTS `user_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `currency` VARCHAR(16) NOT NULL DEFAULT 'USDT',
  `balance` DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_currency` (`user_id`, `currency`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `account_transaction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `transaction_no` VARCHAR(64) NOT NULL,
  `user_id` BIGINT NOT NULL,
  `currency` VARCHAR(16) NOT NULL DEFAULT 'USDT',
  `amount` DECIMAL(18, 2) NOT NULL,
  `type` VARCHAR(32) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `related_type` VARCHAR(32) DEFAULT NULL,
  `related_id` BIGINT DEFAULT NULL,
  `balance_after` DECIMAL(18, 2) DEFAULT NULL,
  `remark` VARCHAR(255) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_transaction_no` (`transaction_no`),
  KEY `idx_user_type_created_at` (`user_id`, `type`, `created_at`),
  KEY `idx_related` (`related_type`, `related_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `campaign_performance_daily` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `campaign_id` BIGINT NOT NULL,
  `stat_date` DATE NOT NULL,
  `participant_count` BIGINT NOT NULL DEFAULT 0,
  `participation_count` BIGINT NOT NULL DEFAULT 0,
  `reward_issued_count` BIGINT NOT NULL DEFAULT 0,
  `reward_issued_amount` DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
  `reward_failed_count` BIGINT NOT NULL DEFAULT 0,
  `currency` VARCHAR(16) NOT NULL DEFAULT 'USDT',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_campaign_date` (`campaign_id`, `stat_date`),
  KEY `idx_campaign_date` (`campaign_id`, `stat_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
