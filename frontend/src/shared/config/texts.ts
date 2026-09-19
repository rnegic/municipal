export const commonTexts = {
  appName: 'УК ↔ Жилец',
  actions: {
    back: 'Назад',
    goHome: 'Вернуться на главную',
    tryAgain: 'Попробовать снова',
    close: 'Закрыть',
    changeAddress: 'Сменить адрес',
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
  },
  underConstruction: {
    title: 'Экран в разработке'
  },
  notFound: {
    title: 'Экран не найден',
    description: 'Похоже, ссылка устарела или в адресе опечатка.',
  },
  a11y: {
    loadingApp: 'Загрузка приложения',
  },
} as const
