INSERT INTO organization (id, name) VALUES ('uk-1', 'УК Наш Дом')
  ON CONFLICT (id) DO NOTHING;
-- fias_id: реальный house_fias_id демо-адреса из DaData (см. docs/uk-integration.md, «Демо-дом»).
INSERT INTO house (id, fias_id, address, organization_id)
  VALUES ('h-1', 'REPLACE-WITH-DADATA-HOUSE-FIAS-ID', 'г Казань, ул Баумана, д 10', 'uk-1')
  ON CONFLICT (id) DO UPDATE SET fias_id = EXCLUDED.fias_id, address = EXCLUDED.address;
INSERT INTO dispatcher (login, password_sha256, organization_id)
  VALUES ('dispatcher', '0ead2060b65992dca4769af601a1b3a35ef38cfad2c2c465bb160ea764157c5d', 'uk-1')
  ON CONFLICT (login) DO NOTHING;
