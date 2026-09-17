export const incidentTexts = {
  focus: {
    sectionTitle: 'Аварии в вашем доме',
    houseOkTitle: 'Всё работает штатно',
    houseOkDescription: 'Активных аварий и отключений по вашему адресу нет',
    headerStatusAlert: 'Есть активная авария',
    joinAction: 'У меня тоже (Подписаться)',
    joinedAction: 'Вы подписались',
    affectedCount: (count: number) => `Уже подписались: ${count}`,
    dueAt: (time: string) => `Планируют устранить к ${time}`,
  },
  severity: {
    critical: 'Авария',
    warning: 'Плановые работы',
  },
  statuses: {
    accepted: 'Принята',
    in_progress: 'В работе',
    verifying: 'Проверка',
    done: 'Закрыта',
  },
  requests: {
    sectionTitle: 'Мои заявки',
    emptyTitle: 'Заявок пока нет',
    emptyDescription: 'Расскажите о проблеме — УК увидит обращение и возьмёт его в работу',
    createdAt: (dateTime: string) => `Отправлена ${dateTime}`,
    dueAt: (time: string) => `Срок — до ${time}`,
  },
  report: {
    action: 'Сообщить о новой проблеме',
  },
  a11y: {
    statusTrack: 'Стадия работ по заявке',
    severityIcon: 'Признак аварии',
  },
} as const
