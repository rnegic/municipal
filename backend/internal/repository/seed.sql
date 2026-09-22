-- Демо-дом и мок-заявки для дефолтного адреса (см. docs/uk-integration.md «Демо-дом»).
-- house_fias_id взят из DaData suggest для "Казань, ул. Баумана, д. 7/10" — держите его в
-- синхроне с тем же id в mock/uk/internal/store/seed.sql.
INSERT INTO uk (external_id, name) VALUES ('uk-1', 'УК Наш Дом')
  ON CONFLICT (external_id) DO NOTHING;

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

-- Третья строка — «объявление»: своей сущности под него нет (event/event_response снесены
-- при мерже uk-contour, см. docs/superpowers/), это обычная заявка с соответствующим title.
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
