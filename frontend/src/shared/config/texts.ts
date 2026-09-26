export const commonTexts = {
  appName: 'УК ↔ Жилец',
  actions: {
    back: 'Назад',
    goHome: 'Вернуться на главную',
    tryAgain: 'Попробовать снова',
    close: 'Закрыть',
    changeAddress: 'Сменить адрес',
    signOut: 'Выйти из аккаунта',
  },
  states: {
    loading: 'Загрузка…',
  },
  errors: {
    unexpectedTitle: 'Что-то пошло не так',
    unexpectedDescription: 'Мы уже знаем о проблеме. Обновите экран или попробуйте позже.',
    networkTitle: 'Нет связи с сервером',
    networkDescription: 'Проверьте подключение и попробуйте ещё раз',
    unauthorizedTitle: 'Не удалось подтвердить вход',
    unauthorizedDescription: 'Откройте приложение из MAX и попробуйте снова',
    forbiddenTitle: 'Недостаточно прав',
    forbiddenDescription: 'Действие доступно только сотруднику УК',
    notFoundTitle: 'Данные не найдены',
    notFoundDescription: 'Возможно, запись удалили или ссылка устарела',
    validationTitle: 'Проверьте данные',
    validationDescription: 'Заполните поля и попробуйте снова',
    businessTitle: 'Не получилось выполнить действие',
    businessDescription: 'Попробуйте ещё раз или обратитесь в УК',
    rateLimitTitle: 'Слишком много заявок',
    rateLimitDescription: 'Попробуйте позже',
  },
  underConstruction: {
    title: 'Экран в разработке'
  },
  select: {
    placeholder: 'Выберите значение',
  },
  photoPicker: {
    add: 'Добавить фото',
    addMore: 'Добавить ещё фото',
    remove: 'Удалить фото',
    limitReached: 'Достигнут лимит фото',
    hint: (maxFiles: number, maxMegabytes: number) =>
      `JPEG, PNG или HEIC · до ${maxFiles} фото · не больше ${maxMegabytes} МБ каждое`,
    typeError: 'Поддерживаются только изображения JPEG, PNG и HEIC',
    sizeError: (maxMegabytes: number) => `Размер фото не должен превышать ${maxMegabytes} МБ`,
    countError: (maxFiles: number) => `Можно приложить не больше ${maxFiles} фото`,
  },
  notFound: {
    title: 'Экран не найден',
    description: 'Похоже, ссылка устарела или в адресе опечатка.',
  },
  a11y: {
    loadingApp: 'Загрузка приложения',
    photoList: 'Приложенные фото',
    photoPreview: (index: number) => `Фото ${index}`,
  },
} as const
