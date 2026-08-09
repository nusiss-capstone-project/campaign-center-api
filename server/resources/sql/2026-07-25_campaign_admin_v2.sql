-- Campaign admin v2 schema (drafts / flattened reward rules / ledger).
-- For existing DBs: run this migration.
-- Greenfield installs should also apply the matching definition in 2026-05-10_schema_init.sql
-- (or re-run init then this file for ALTER-only environments).

-- ---------------------------------------------------------------------------
-- campaigns: reshape for admin v2
-- ---------------------------------------------------------------------------
ALTER TABLE `campaigns`
  DROP INDEX `idx_campaigns_status_type`,
  DROP COLUMN `type`,
  DROP COLUMN `reward_rules`,
  CHANGE COLUMN `target_market` `market` VARCHAR(64) DEFAULT NULL,
  DROP COLUMN `target_user_segment`,
  ADD COLUMN `target_user_group_id` BIGINT NOT NULL DEFAULT 0 AFTER `campaign_end_time`,
  ADD COLUMN `budget_project_id` BIGINT NOT NULL DEFAULT 0 AFTER `target_user_group_id`,
  ADD COLUMN `time_zone` VARCHAR(64) NOT NULL DEFAULT '' AFTER `landing_page_id`,
  ADD KEY `idx_campaigns_status` (`status`);

ALTER TABLE `campaigns`
  MODIFY COLUMN `status` SMALLINT DEFAULT NULL COMMENT '1: draft 2: published';

-- ---------------------------------------------------------------------------
-- campaign_drafts: versioned editable JSON body
-- status: draft | published
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
-- campaign_reward_rules: flattened task / task_group reward bindings
-- ref_client: task | task_group
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
-- campaign_user_rewards_ledger: user reward distribution ledger
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
