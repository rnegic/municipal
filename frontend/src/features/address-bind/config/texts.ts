export const addressBindTexts = {
  title: 'Укажите адрес дома',
  description: 'Покажем аварии и заявки именно по вашему дому',
  label: 'Адрес дома',
  placeholder: 'Например, Казань, ул. Баумана, д 7/10',
  submit: 'Сохранить адрес',
  validation: 'Введите улицу и номер дома',
  suggestionsLabel: 'Подсказки адреса',
  suggestionsLoading: 'Ищем адрес…',
  suggestionsEmpty: 'Ничего не нашли. Уточните улицу и номер дома',
  locate: 'Определить адрес по геолокации',
  locationEmpty: 'Рядом не нашли дом. Введите адрес вручную',
  locationErrors: {
    unsupported: 'Геолокация недоступна на этом устройстве',
    denied: 'Разрешите доступ к геолокации',
    unavailable: 'Не удалось определить местоположение',
    timeout: 'Не удалось определить местоположение',
  },
} as const