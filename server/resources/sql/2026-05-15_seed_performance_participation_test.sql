-- Dev seed: admin performance & participations APIs
-- Campaign ID: 10001
--
-- After running, expected:
--   GET .../campaigns/10001/performance/summary
--     participantCount=4, participationCount=4,
--     rewardIssuedCount=2, rewardIssuedAmount=20.00
--
--   GET .../campaigns/10001/performance/daily?startDate=2026-05-13&endDate=2026-05-15
--     3 rows (2026-05-13 / 14 / 15)
--
--   GET .../campaigns/10001/participations?page=1&pageSize=20
--     total=4
--   GET .../participations?status=GRANTED  -> 2 rows (user 10001, 10002)
--   GET .../participations?userId=10003    -> 1 row (PENDING_REVIEW)

USE `campaign_center`;

-- ---------------------------------------------------------------------------
-- Cleanup (re-runnable)
-- ---------------------------------------------------------------------------
DELETE FROM `campaign_performance_daily` WHERE `campaign_id` = 10001;
DELETE FROM `campaign_participants`      WHERE `campaign_id` = 10001;
DELETE FROM `reward_transactions`        WHERE `campaign_id` = 10001;
DELETE FROM `campaigns`                  WHERE `id` = 10001;
DELETE FROM `campaign_landing_pages`     WHERE `id` = 20001;
DELETE FROM `users`                      WHERE `id` IN (10001, 10002, 10003, 10004);

-- ---------------------------------------------------------------------------
-- Landing page + campaign (published)
-- ---------------------------------------------------------------------------
INSERT INTO `campaign_landing_pages` (
  `id`, `default_lang`, `banner_image_url`, `title`, `description`, `terms`,
  `created_at`, `updated_at`, `status`, `created_by`, `updated_by`
) VALUES (
  20001, 'en-US', 'https://example.com/banner.png',
  'Top up {{threshold}} get {{reward}}',
  'Demo campaign for performance APIs.', 'Terms apply.',
  NOW(3), NOW(3), 2, 'seed', 'seed'
);

INSERT INTO `campaigns` (
  `id`, `name`, `type`, `target_market`,
  `registration_start_time`, `registration_end_time`,
  `campaign_start_time`, `campaign_end_time`,
  `target_user_segment`, `reward_rules`, `status`,
  `created_at`, `updated_at`, `created_by`, `updated_by`, `landing_page_id`
) VALUES (
  10001, 'Perf Demo Top-up Reward', 'TOPUP_REWARD', 'US',
  DATE_SUB(NOW(3), INTERVAL 7 DAY), DATE_ADD(NOW(3), INTERVAL 30 DAY),
  DATE_SUB(NOW(3), INTERVAL 7 DAY), DATE_ADD(NOW(3), INTERVAL 60 DAY),
  'NEW_USER',
  '{"topupThreshold":100,"rewardAmount":10,"rewardType":"BONUS_CREDIT","maxClaimPerUser":1}',
  2, NOW(3), NOW(3), 'seed', 'seed', 20001
);

-- ---------------------------------------------------------------------------
-- Users (optional; participations API does not require users table rows)
-- ---------------------------------------------------------------------------
INSERT INTO `users` (`id`, `name`, `market`, `segment`, `kyc_status`, `risk_level`, `created_at`) VALUES
  (10001, 'Alice', 'US', 'NEW_USER', 'PASSED', 'LOW',  NOW(3)),
  (10002, 'Bob',   'US', 'NEW_USER', 'PASSED', 'LOW',  NOW(3)),
  (10003, 'Carol', 'US', 'NEW_USER', 'PASSED', 'HIGH', NOW(3)),
  (10004, 'Dave',  'US', 'NEW_USER', 'PASSED', 'LOW',  NOW(3));

-- ---------------------------------------------------------------------------
-- Participations (drives summary + participations list)
-- join_status: JOINED | reward_status: GRANTED / PENDING_REVIEW / NOT_GRANTED
-- ---------------------------------------------------------------------------
INSERT INTO `campaign_participants` (
  `id`, `campaign_id`, `user_id`, `join_status`, `task_status`,
  `topup_amount`, `risk_status`, `reward_status`, `reward_amount`,
  `joined_at`, `completed_at`, `rewarded_at`, `updated_at`
) VALUES
  (90001, 10001, 10001, 'JOINED', 'COMPLETED', 120.00, 'APPROVED',      'GRANTED',        10.00,
   '2026-05-13 10:00:00.000', '2026-05-13 10:20:00.000', '2026-05-13 10:21:00.000', NOW(3)),
  (90002, 10001, 10002, 'JOINED', 'COMPLETED', 150.00, 'APPROVED',      'GRANTED',        10.00,
   '2026-05-14 09:00:00.000', '2026-05-14 09:30:00.000', '2026-05-14 09:31:00.000', NOW(3)),
  (90003, 10001, 10003, 'JOINED', 'COMPLETED', 200.00, 'MANUAL_REVIEW', 'PENDING_REVIEW',  0.00,
   '2026-05-15 08:00:00.000', '2026-05-15 08:10:00.000', NULL,                      NOW(3)),
  (90004, 10001, 10004, 'JOINED', 'NOT_STARTED', NULL, NULL,          'NOT_GRANTED',     0.00,
   '2026-05-15 11:00:00.000', NULL, NULL, NOW(3));

-- ---------------------------------------------------------------------------
-- Daily performance (drives /performance/daily only)
-- ---------------------------------------------------------------------------
INSERT INTO `campaign_performance_daily` (
  `campaign_id`, `stat_date`, `participant_count`, `participation_count`,
  `reward_issued_count`, `reward_issued_amount`, `reward_failed_count`,
  `currency`, `created_at`, `updated_at`
) VALUES
  (10001, '2026-05-13', 1, 1, 1, 10.00, 0, 'USDT', NOW(3), NOW(3)),
  (10001, '2026-05-14', 1, 1, 1, 10.00, 0, 'USDT', NOW(3), NOW(3)),
  (10001, '2026-05-15', 2, 2, 0,  0.00, 1, 'USDT', NOW(3), NOW(3));

-- Optional: reward_transactions audit rows (not used by performance APIs today)
INSERT INTO `reward_transactions` (
  `id`, `campaign_id`, `user_id`, `participant_id`,
  `reward_type`, `reward_amount`, `status`, `created_at`
) VALUES
  (91001, 10001, 10001, 90001, 'BONUS_CREDIT', 10.00, 'COMPLETED', '2026-05-13 10:21:00.000'),
  (91002, 10001, 10002, 90002, 'BONUS_CREDIT', 10.00, 'COMPLETED', '2026-05-14 09:31:00.000');
