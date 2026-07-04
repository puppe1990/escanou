-- up
ALTER TABLE price_reports ADD COLUMN disputes INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS price_disputes (
  price_report_id INTEGER NOT NULL REFERENCES price_reports(id),
  user_id INTEGER NOT NULL REFERENCES users(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (price_report_id, user_id)
);

-- down
DROP TABLE IF EXISTS price_disputes;