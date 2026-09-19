import dayjs from 'dayjs'
import 'dayjs/locale/ru'

dayjs.locale('ru')
export const formatTime = (value: string | number | Date): string => dayjs(value).format('HH:mm')

export const formatDateTime = (value: string | number | Date): string =>
  dayjs(value).format('D MMMM, HH:mm')

export const formatDate = (value: string | number | Date): string => dayjs(value).format('DD.MM.YYYY')
