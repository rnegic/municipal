const ruPlural = new Intl.PluralRules('ru')

const affectedForms: Partial<Record<Intl.LDMLPluralRule, string>> = {
  one: 'житель подтвердил',
  few: 'жителя подтвердили',
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
  },
} as const
