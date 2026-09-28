INSERT INTO uk (external_id, name, inn, ogrn, license_number, license_valid_until,
                phone, emergency_phone, email, website, office_address, working_hours)
  VALUES ('uk-1', 'УК Наш Дом', '1655000003', '1021602000000', '16-000123', '2030-12-31',
          '+7 (843) 200-00-00', '+7 (843) 200-01-01', 'info@uk-nash-dom.ru', 'https://uk-nash-dom.ru',
          'г Казань, ул Баумана, д 7/10, офис 101', 'Пн–Пт 8:00–17:00, обед 12:00–13:00')
  ON CONFLICT (external_id) DO UPDATE SET
    inn = EXCLUDED.inn, ogrn = EXCLUDED.ogrn,
    license_number = EXCLUDED.license_number, license_valid_until = EXCLUDED.license_valid_until,
    phone = COALESCE(EXCLUDED.phone, uk.phone), emergency_phone = COALESCE(EXCLUDED.emergency_phone, uk.emergency_phone),
    email = COALESCE(EXCLUDED.email, uk.email), website = COALESCE(EXCLUDED.website, uk.website),
    office_address = COALESCE(EXCLUDED.office_address, uk.office_address), working_hours = COALESCE(EXCLUDED.working_hours, uk.working_hours);

INSERT INTO app_user (full_name, role, uk_id, position, password_hash, ads_authority)
  SELECT 'Ильдар Хайруллин', 'uk_dispatcher', uk.id, 'Диспетчер АДС',
         '$2a$10$UDcFWb4lzy5Y89wpfzu.uOcQ0RI.vS0bLSrechwEM2GRocAVpFUka', true
  FROM uk WHERE uk.external_id = 'uk-1'
    AND NOT EXISTS (SELECT 1 FROM app_user a WHERE a.uk_id = uk.id AND a.role = 'uk_dispatcher');

INSERT INTO house (address_raw, house_fias_id, uk_id, external_id)
  SELECT 'Казань, ул. Баумана, д. 7/10', 'c6f16fa1-aa2d-4507-bda5-d0ab2cc7a531', uk.id, 'h-1'
  FROM uk WHERE uk.external_id = 'uk-1'
  ON CONFLICT (house_fias_id) DO UPDATE SET
    address_raw = EXCLUDED.address_raw, uk_id = EXCLUDED.uk_id, external_id = EXCLUDED.external_id;

INSERT INTO app_user (max_user_id, full_name, house_id)
  SELECT v.max_user_id, v.full_name, h.id
  FROM house h, (VALUES
    (900000001, 'Ирина Демо'),
    (900000002, 'Пётр Демо'),
    (900000003, 'Светлана Демо')
  ) AS v(max_user_id, full_name)
  WHERE h.house_fias_id = 'c6f16fa1-aa2d-4507-bda5-d0ab2cc7a531'
  ON CONFLICT (max_user_id) DO NOTHING;

INSERT INTO incident (house_id, title, severity, reporter_id, description, entrance, riser, status)
  SELECT h.id, v.title, v.severity, u.id, v.description, v.entrance, v.riser, 'accepted'
  FROM house h
  CROSS JOIN (VALUES
    (900000001, 'Течёт крыша над 3 подъездом', 'critical',
     'Второй день капает с потолка на лестничной клетке.', '3', NULL),
    (900000002, 'Не горит свет в подъезде', 'warning',
     'На 2 и 3 этажах не работает освещение с вечера.', '1', NULL),
    (900000003, 'Объявление: отключение горячей воды 25–27 сентября', 'warning',
     'Плановые работы на теплосети, УК «Наш Дом» просит заранее набрать воду.', NULL, NULL)
  ) AS v(reporter, title, severity, description, entrance, riser)
  JOIN app_user u ON u.max_user_id = v.reporter
  WHERE h.house_fias_id = 'c6f16fa1-aa2d-4507-bda5-d0ab2cc7a531'
    AND NOT EXISTS (SELECT 1 FROM incident i WHERE i.house_id = h.id AND i.title = v.title);
