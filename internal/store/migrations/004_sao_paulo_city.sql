-- up
UPDATE user_profiles SET city = 'São Paulo' WHERE city = 'Curitiba';

DELETE FROM price_confirmations
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN supermarkets sm ON sm.id = pr.supermarket_id
  WHERE sm.name IN (
    'Angeloni Bacacheri',
    'Bistek Hauer',
    'Carrefour Portão',
    'Condor Água Verde'
  )
);

DELETE FROM price_reports
WHERE supermarket_id IN (
  SELECT id FROM supermarkets WHERE name IN (
    'Angeloni Bacacheri',
    'Bistek Hauer',
    'Carrefour Portão',
    'Condor Água Verde'
  )
);

DELETE FROM supermarkets WHERE name IN (
  'Angeloni Bacacheri',
  'Bistek Hauer',
  'Carrefour Portão',
  'Condor Água Verde'
);

-- down
-- Data migration only; SP supermarkets are re-seeded on boot in development.