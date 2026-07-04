-- up
INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Extra Belenzinho', 'R. Oriente, 234 — Belenzinho', -23.5398, -46.5865
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Extra Belenzinho');

INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Assaí Belém', 'Av. Alcântara Machado, 664 — Belém', -23.5368, -46.5832
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Assaí Belém');

INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Dia Belenzinho', 'R. José Belenzinho, 289 — Belenzinho', -23.5442, -46.5948
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Dia Belenzinho');

INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Extra Tatuapé', 'R. Serra de Bragança, 1555 — Tatuapé', -23.5492, -46.5418
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Extra Tatuapé');

INSERT INTO supermarkets (name, address, lat, lng)
SELECT 'Assaí Tatuapé', 'R. Maria Cândida, 1899 — Tatuapé', -23.5496, -46.5392
WHERE NOT EXISTS (SELECT 1 FROM supermarkets WHERE name = 'Assaí Tatuapé');

UPDATE supermarkets
SET address = 'R. Serra de Bragança, 629 — Tatuapé'
WHERE name = 'Carrefour Tatuapé' AND address NOT LIKE '%Tatuapé%';

-- down
DELETE FROM price_confirmations
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN supermarkets sm ON sm.id = pr.supermarket_id
  WHERE sm.name IN (
    'Extra Belenzinho', 'Assaí Belém', 'Dia Belenzinho', 'Extra Tatuapé', 'Assaí Tatuapé'
  )
);

DELETE FROM price_disputes
WHERE price_report_id IN (
  SELECT pr.id FROM price_reports pr
  JOIN supermarkets sm ON sm.id = pr.supermarket_id
  WHERE sm.name IN (
    'Extra Belenzinho', 'Assaí Belém', 'Dia Belenzinho', 'Extra Tatuapé', 'Assaí Tatuapé'
  )
);

DELETE FROM price_reports
WHERE supermarket_id IN (
  SELECT id FROM supermarkets WHERE name IN (
    'Extra Belenzinho', 'Assaí Belém', 'Dia Belenzinho', 'Extra Tatuapé', 'Assaí Tatuapé'
  )
);

DELETE FROM supermarkets WHERE name IN (
  'Extra Belenzinho', 'Assaí Belém', 'Dia Belenzinho', 'Extra Tatuapé', 'Assaí Tatuapé'
);