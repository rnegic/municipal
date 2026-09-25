const ruPlural = new Intl.PluralRules('ru')

const affectedForms: Partial<Record<Intl.LDMLPluralRule, string>> = {
  one: 'житель подтвердил',
  few: 'жителя подтвердили',
}

const requestForms: Partial<Record<Intl.LDMLPluralRule, string>> = {
  one: 'заявку',
  few: 'заявки',
}

const duplicateForms: Partial<Record<Intl.LDMLPluralRule, string>> = {
  one: 'дубликат',
  few: 'дубликата',
}

export const incidentTexts = {
  focus: {
    sectionTitle: 'Аварии в вашем доме',
    houseOkTitle: 'Всё работает штатно',
    houseOkEmptyTitle: 'В вашем доме всё работает штатно',
    houseOkEmptyDescription: 'УК на страже',
    headerStatusAlert: 'Есть активная авария',
    joinAction: 'У меня тоже (Подписаться)',
    joinedAction: 'Вы подписались',
    affectedCount: (count: number) =>
      `+${count} ${affectedForms[ruPlural.select(count)] ?? 'жителей подтвердили'}`,
    viewAllAction: (count: number) => `Все заявки · ${count}`,
  },
  houseIncidents: {
    title: 'Текущие заявки дома',
    subtitle: 'Все активные обращения по вашему адресу',
    emptyTitle: 'Активных заявок нет',
    emptyDescription: 'Как только в доме появится авария, она окажется здесь',
  },
  merge: {
    startAction: 'Объединить заявки',
    cancelAction: 'Отменить выбор',
    submitAction: (count: number) =>
      `Объединить ${count} ${requestForms[ruPlural.select(count)] ?? 'заявок'}`,
    selectionHint: 'Отметьте дубликаты одной проблемы — минимум две заявки одного дома',
    selectedCount: (count: number) => `Выбрано: ${count}`,
    sheetTitle: 'Объединение заявок',
    sheetDescription: 'Выберите главную заявку — в неё перейдут подписанты и фото остальных',
    targetLegend: 'Главная заявка',
    duplicatesBadge: (count: number) =>
      `${count} ${duplicateForms[ruPlural.select(count)] ?? 'дубликатов'}`,
    blockers: {
      tooFew: 'Отметьте минимум две заявки',
      closed: 'Закрытые заявки объединять нельзя',
      differentHouses: 'Объединять можно только заявки одного дома',
    },
  },
  queue: {
    priorityBadge: 'Приоритет',
    signatories: 'Подписантов',
    reporter: 'Заявитель',
    due: 'Срок',
    createdAt: 'Создано',
    sortHint: 'Сначала заявки с наибольшим числом подписантов',
    selectCard: 'Выбрать заявку для объединения',
  },
  severity: {
    critical: 'Авария',
    warning: 'Плановые работы',
  },
  statuses: {
    accepted: 'Принята',
    in_progress: 'В работе',
    verifying: 'Проверка жителями',
    done: 'Решена',
  },
  confirm: {
    action: 'Подтвердите',
    done: 'Спасибо, подтверждено',
  },
  requests: {
    sectionTitle: 'Мои заявки',
    emptyTitle: 'Вы еще не сообщали о проблемах',
    emptyDescription: 'Расскажите о проблеме — УК увидит обращение и возьмёт его в работу',
    createdAt: (dateTime: string) => `Отправлена ${dateTime}`,
    dueAt: (time: string) => `Срок — до ${time}`,
  },
  report: {
    action: 'Сообщить о новой проблеме',
  },
  categories: {
    WATER_HEAT: 'Водоснабжение и отопление',
    ELECTRICITY: 'Электричество',
    ELEVATOR: 'Лифт',
    CLEANING_YARD: 'Уборка, двор, подъезд, вандализм',
    BUILDING_STRUCTURE: 'Конструктив здания, крыша, фасад',
    CITY_TERRITORY: 'Городская территория',
  },
  authorities: {
    UK: {
      name: 'Управляющая компания',
      scope: 'Подъезд, двор, лифт, стояки до вентиля',
    },
    FKR: {
      name: 'Фонд капитального ремонта',
      scope: 'Протекающая крыша, трещины в несущих стенах, фасад старого дома',
    },
    RSO: {
      name: 'Ресурсоснабжающая организация',
      scope: 'Отключения воды, тепла и света на уровне района',
    },
    MUNICIPALITY: {
      name: 'Муниципалитет',
      scope: 'Дороги за пределами двора, открытые люки на проезжей части',
    },
    OWNER: {
      name: 'Собственник',
      scope: 'Трубы внутри квартиры после счётчика или вентиля',
    },
  },
  a11y: {
    status: 'Текущий статус заявки',
    severityIcon: 'Признак аварии',
    supporters: 'Жители, подтвердившие проблему',
  },
} as const
