-- up
-- Reports flagged by the old single-click downvote (before dispute threshold).
UPDATE price_reports
SET flagged = 0
WHERE flagged = 1 AND COALESCE(disputes, 0) < 3;

-- down
-- no-op: cannot restore which reports were incorrectly flagged