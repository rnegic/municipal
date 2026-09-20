import dayjs from 'dayjs'
import 'dayjs/locale/ru'

dayjs.locale('ru')

export const formatEventSchedule = (from: string, to: string): string => {
  const start = dayjs(from)
  const end = dayjs(to)

  if (!start.isSame(end, 'day')) {
    return `${start.format('D MMMM, HH:mm')} — ${end.format('D MMMM, HH:mm')}`
  }

  const today = dayjs()
  const day = start.isSame(today, 'day')
    ? 'Сегодня'
    : start.isSame(today.add(1, 'day'), 'day')
      ? 'Завтра'
      : start.format('D MMMM')

  return `${day} с ${start.format('HH:mm')} до ${end.format('HH:mm')}`
}
