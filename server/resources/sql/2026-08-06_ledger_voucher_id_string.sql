-- Alter campaign_user_rewards_ledger.voucher_id to string for reward-mservice voucher ids.
ALTER TABLE `campaign_user_rewards_ledger`
  MODIFY COLUMN `voucher_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'reward voucher id';

-- Idempotent unique key for (user, campaign, rule).
SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'campaign_user_rewards_ledger'
    AND INDEX_NAME = 'uk_ledger_user_campaign_rule'
);
SET @sql := IF(@has_uk = 0,
  'ALTER TABLE `campaign_user_rewards_ledger` ADD UNIQUE KEY `uk_ledger_user_campaign_rule` (`user_id`, `campaign_id`, `rule_id`)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
