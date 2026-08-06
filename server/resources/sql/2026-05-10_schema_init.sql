-- Phase1 User Top-up Campaign — database schema init
-- Generated to align with server/repository/mysql/model and design doc.
-- Target: MySQL 8.x, utf8mb4

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE DATABASE IF NOT EXISTS `campaign_center`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE `campaign_center`;

-- ---------------------------------------------------------------------------
-- campaign_landing_pages (referenced by campaigns.landing_page_id)
-- status: 1 draft, 2 published, 3 archive
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `campaign_landing_pages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `default_lang` VARCHAR(16) DEFAULT 'en',
  `banner_image_url` VARCHAR(512) DEFAULT NULL,
  `title` VARCHAR(255) DEFAULT NULL,
  `description` TEXT,
  `terms` TEXT,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `status` SMALLINT DEFAULT NULL COMMENT '1: draft 2: published 3: archive',
  `created_by` VARCHAR(255) NOT NULL DEFAULT '',
  `updated_by` VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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

-- ---------------------------------------------------------------------------
-- users (mock profile for KYC / segment / risk checks)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) DEFAULT NULL,
  `market` VARCHAR(64) DEFAULT NULL,
  `segment` VARCHAR(64) DEFAULT NULL,
  `kyc_status` VARCHAR(32) DEFAULT NULL,
  `risk_level` VARCHAR(32) DEFAULT NULL,
  `created_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- campaigns (admin v2)
-- status: 1 draft, 2 published
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `campaigns` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) DEFAULT NULL,
  `market` VARCHAR(64) DEFAULT NULL,
  `registration_start_time` DATETIME(3) DEFAULT NULL,
  `registration_end_time` DATETIME(3) DEFAULT NULL,
  `campaign_start_time` DATETIME(3) DEFAULT NULL,
  `campaign_end_time` DATETIME(3) DEFAULT NULL,
  `target_user_group_id` BIGINT NOT NULL DEFAULT 0,
  `budget_project_id` BIGINT NOT NULL DEFAULT 0,
  `status` SMALLINT DEFAULT NULL COMMENT '1: draft 2: published',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  `created_by` VARCHAR(255) NOT NULL DEFAULT '',
  `updated_by` VARCHAR(255) NOT NULL DEFAULT '',
  `landing_page_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'bind campaign_landing_pages.id',
  `time_zone` VARCHAR(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `idx_campaigns_landing_page_id` (`landing_page_id`),
  KEY `idx_campaigns_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- campaign_drafts
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `campaign_drafts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `activity_id` BIGINT NOT NULL COMMENT 'campaigns.id',
  `content` TEXT COMMENT 'JSON body of editable campaign fields',
  `version` INT NOT NULL DEFAULT 1,
  `status` VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT 'draft / published',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_campaign_drafts_activity_version` (`activity_id`, `version`),
  KEY `idx_campaign_drafts_activity_status` (`activity_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- campaign_reward_rules
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `campaign_reward_rules` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `campaign_id` BIGINT NOT NULL,
  `ref_client` VARCHAR(32) NOT NULL COMMENT 'task / task_group',
  `ref_id` BIGINT NOT NULL COMMENT 'task_id / task_group_id',
  `reward_template_id` BIGINT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_campaign_reward_rules_campaign` (`campaign_id`),
  KEY `idx_campaign_reward_rules_ref` (`ref_client`, `ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- campaign_user_rewards_ledger
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `campaign_user_rewards_ledger` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `campaign_id` BIGINT NOT NULL,
  `rule_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'campaign_reward_rules.id',
  `reward_status` VARCHAR(64) NOT NULL DEFAULT 'pending_distribution'
    COMMENT 'pending_distribution / distributing / distribute_success / distribute_fail',
  `voucher_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'reward voucher id',
  `created_at` DATETIME(3) DEFAULT NULL,
  `updated_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_rewards_ledger_user` (`user_id`),
  KEY `idx_user_rewards_ledger_campaign` (`campaign_id`),
  KEY `idx_user_rewards_ledger_rule` (`rule_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- campaign_participants
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS CREATE TABLE `campaign_participants` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `campaign_id` bigint DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `join_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `task_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `topup_amount` decimal(18,2) DEFAULT NULL,
  `risk_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reward_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `reward_amount` decimal(18,2) DEFAULT NULL,
  `joined_at` datetime(3) DEFAULT NULL,
  `completed_at` datetime(3) DEFAULT NULL,
  `rewarded_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user_campaign` (`user_id`,`campaign_id`),
  KEY `idx_participant_campaign` (`campaign_id`),
  KEY `idx_participant_user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=90007 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci

-- ---------------------------------------------------------------------------
-- reward_transactions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `reward_transactions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `campaign_id` BIGINT DEFAULT NULL,
  `user_id` BIGINT DEFAULT NULL,
  `participant_id` BIGINT DEFAULT NULL,
  `reward_type` VARCHAR(64) DEFAULT NULL,
  `reward_amount` DECIMAL(18,2) DEFAULT NULL,
  `status` VARCHAR(32) DEFAULT NULL,
  `reason` VARCHAR(255) DEFAULT NULL,
  `created_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_reward_txn_campaign` (`campaign_id`),
  KEY `idx_reward_txn_user` (`user_id`),
  KEY `idx_reward_txn_participant` (`participant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ---------------------------------------------------------------------------
-- audit_logs
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `entity_type` VARCHAR(64) DEFAULT NULL,
  `entity_id` BIGINT DEFAULT NULL,
  `action` VARCHAR(64) DEFAULT NULL,
  `operator_name` VARCHAR(64) DEFAULT NULL,
  `detail_json` TEXT,
  `created_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_audit_entity` (`entity_type`, `entity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;

-- ---------------------------------------------------------------------------
-- Optional dev seed (uncomment to load sample rows; adjust IDs as needed)
-- ---------------------------------------------------------------------------
-- INSERT INTO `campaign_landing_pages` (`id`, `default_lang`, `banner_image_url`, `title`, `description`, `terms`, `created_at`, `updated_at`, `status`, `created_by`, `updated_by`)
-- VALUES
--   (2001, 'en-US', 'https://example.com/banner.png', 'Top up {{threshold}} and get {{reward}} bonus', 'Join the campaign.', 'Terms apply.', NOW(3), NOW(3), 2, 'seed', 'seed');
--
-- INSERT INTO `campaigns` (`id`, `name`, `type`, `target_market`, `registration_start_time`, `registration_end_time`, `campaign_start_time`, `campaign_end_time`, `target_user_segment`, `reward_rules`, `status`, `created_at`, `updated_at`, `created_by`, `updated_by`, `landing_page_id`)
-- VALUES
--   (1001, 'New User Top-up Reward', 'TOPUP_REWARD', 'US', NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3), DATE_ADD(NOW(3), INTERVAL 60 DAY), 'NEW_USER',
--    '{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}', 2, NOW(3), NOW(3), 'seed', 'seed', 2001);
--
-- INSERT INTO `users` (`id`, `name`, `market`, `segment`, `kyc_status`, `risk_level`, `created_at`)
-- VALUES
--   (1001, 'Demo User', 'US', 'NEW_USER', 'PASSED', 'LOW', NOW(3)),
--   (1003, 'High Risk User', 'US', 'NEW_USER', 'PASSED', 'HIGH', NOW(3));
