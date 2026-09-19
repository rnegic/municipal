export const incidentTexts = {
  focus: {
    sectionTitle: 'Аварии в вашем доме',
    houseOkTitle: 'Всё работает штатно',
    houseOkDescription: 'Активных аварий и отключений по вашему адресу нет',
    headerStatusAlert: 'Есть активная авария',
    joinAction: 'У меня тоже (Подписаться)',
    joinedAction: 'Вы подписались',
    affectedCount: (count: number) => `+${count} жителей подтвердили`,
  },
  severity: {
    critical: 'Авария',
    warning: 'Плановые работы',
  },
  statuses: {
    accepted: 'Принята',
    in_progress: 'В работе',
    verifying: 'Решена',
    done: 'Решена',
  },
  confirm: {
    action: 'Подтвердите',
    done: 'Спасибо, подтверждено',
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
    status: 'Текущий статус заявки',
    severityIcon: 'Признак аварии',
  },
} as const
