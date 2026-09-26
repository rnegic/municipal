INSERT INTO organization (id, name, phone, emergency_phone, email, website, office_address, working_hours)
  VALUES ('uk-1', 'УК Наш Дом', '+7 (843) 200-00-00', '+7 (843) 200-01-01', 'info@uk-nash-dom.ru',
          'https://uk-nash-dom.ru', 'г Казань, ул Баумана, д 7/10, офис 101', 'Пн–Пт 8:00–17:00, обед 12:00–13:00')
  ON CONFLICT (id) DO UPDATE SET
    phone = EXCLUDED.phone, emergency_phone = EXCLUDED.emergency_phone, email = EXCLUDED.email,
    website = EXCLUDED.website, office_address = EXCLUDED.office_address, working_hours = EXCLUDED.working_hours;
-- fias_id: реальный house_fias_id демо-адреса из DaData (см. docs/uk-integration.md, «Демо-дом»).
INSERT INTO house (id, fias_id, address, organization_id)
  VALUES ('h-1', 'c6f16fa1-aa2d-4507-bda5-d0ab2cc7a531', 'г Казань, ул Баумана, д 7/10', 'uk-1')
  ON CONFLICT (id) DO UPDATE SET fias_id = EXCLUDED.fias_id, address = EXCLUDED.address;
INSERT INTO dispatcher (login, password_sha256, organization_id)
  VALUES ('dispatcher', '0ead2060b65992dca4769af601a1b3a35ef38cfad2c2c465bb160ea764157c5d', 'uk-1')
  ON CONFLICT (login) DO NOTHING;
