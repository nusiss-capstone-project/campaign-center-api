-- Simplify campaign_participants to join-only fields.
-- Target shape: id, campaign_id, user_id (unique), join_at, created_at, updated_at

CREATE TABLE IF NOT EXISTS `campaign_participants_new` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `campaign_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `join_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL,
  `updated_at` DATETIME(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_participant_campaign_user` (`campaign_id`, `user_id`),
  KEY `idx_participant_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `campaign_participants_new` (
  `id`, `campaign_id`, `user_id`, `join_at`, `created_at`, `updated_at`
)
SELECT
  `id`,
  `campaign_id`,
  `user_id`,
  COALESCE(`joined_at`, `updated_at`, CURRENT_TIMESTAMP(3)),
  COALESCE(`joined_at`, `updated_at`, CURRENT_TIMESTAMP(3)),
  COALESCE(`updated_at`, CURRENT_TIMESTAMP(3))
FROM `campaign_participants`
WHERE `campaign_id` IS NOT NULL AND `user_id` IS NOT NULL;

DROP TABLE `campaign_participants`;
RENAME TABLE `campaign_participants_new` TO `campaign_participants`;
