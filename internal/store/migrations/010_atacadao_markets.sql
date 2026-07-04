-- up
INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Atacadão Belenzinho', 'Av. do Estado, 5533 — Brás', -23.5342, -46.5778
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Atacadão Belenzinho');

INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Atacadão Tatuapé', 'R. Maria Cândida, 1763 — Tatuapé', -23.5488, -46.5378
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Atacadão Tatuapé');

-- down
DELETE FROM price_confirmations
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN supermarkets sm ON sm.id = pr.supermarket_id
  WHERE sm.name IN ('Atacadão Belenzinho', 'Atacadão Tatuapé')
);

DELETE FROM price_disputes
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN supermarkets sm ON sm.id = pr.supermarket_id
  WHERE sm.name IN ('Atacadão Belenzinho', 'Atacadão Tatuapé')
);

DELETE FROM price_reports
WHERE supermarket_id IN (
  SELECT id FROM supermarkets WHERE name IN ('Atacadão Belenzinho', 'Atacadão Tatuapé')
);

DELETE FROM supermarkets WHERE name IN ('Atacadão Belenzinho', 'Atacadão Tatuapé');