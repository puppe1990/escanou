-- up
ALTER TABLE products ADD COLUMN brand TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN quantity TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN image_url TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN source TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE products ADD COLUMN off_fetched_at DATETIME;

-- down
-- SQLite cannot drop columns easily; recreate would lose data. No down migration.