-- up
UPDATE user_profiles SET city = 'São Paulo' WHERE city = 'Curitiba';

DELETE FROM supermarkets WHERE name IN (
  'Angeloni Bacacheri',
  'Bistek Hauer',
  'Carrefour Portão',
  'Condor Água Verde'
);

-- down
-- Data migration only; SP supermarkets are re-seeded on boot in development.