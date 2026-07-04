-- up
DELETE FROM price_confirmations
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN users u ON u.id = pr.user_id
  WHERE u.email = 'demo@example.com'
);

DELETE FROM price_reports
WHERE user_id IN (SELECT id FROM users WHERE email = 'demo@example.com');

UPDATE user_profiles
SET points = 0, display_name = 'Demo', city = 'São Paulo'
WHERE user_id IN (SELECT id FROM users WHERE email = 'demo@example.com');

DELETE FROM user_badges
WHERE user_id IN (SELECT id FROM users WHERE email = 'demo@example.com');

-- down
-- Demo reports are not re-seeded; scan to rebuild history.