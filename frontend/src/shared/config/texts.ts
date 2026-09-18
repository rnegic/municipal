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
