import dayjs from 'dayjs'

import {
  incidentSchema,
  residentRequestSchema,
  type Incident,
  type ResidentRequest,
} from './schema'

const HOUSE_ID = 'house-1'

export const PLACEHOLDER_ACTIVE_INCIDENT: Incident | null = incidentSchema.parse({
  id: 'incident-1',
  houseId: HOUSE_ID,
  title: 'Нет горячей воды',
  description: 'Мастер работает. Починят к 15:00',
  severity: 'critical',
  status: 'in_progress',
  affectedCount: 12,
  createdAt: dayjs().subtract(2, 'hour').toISOString(),
  dueAt: dayjs().hour(15).minute(0).second(0).millisecond(0).toISOString(),
})

export const PLACEHOLDER_REQUESTS: ResidentRequest[] = residentRequestSchema.array().parse([
  {
    id: 'request-1',
    title: 'Течёт кран на кухне',
    status: 'verifying',
    createdAt: dayjs().subtract(5, 'hour').toISOString(),
    dueAt: dayjs().add(1, 'day').toISOString(),
  },
  {
    id: 'request-2',
    title: 'Не закрывается доводчик в подъезде',
    status: 'accepted',
    createdAt: dayjs().subtract(1, 'day').toISOString(),
    dueAt: dayjs().add(3, 'day').toISOString(),
  },
  {
    id: 'request-3',
    title: 'Перегорела лампа на лестничной клетке',
    status: 'done',
    createdAt: dayjs().subtract(12, 'day').toISOString(),
    dueAt: null,
  },
])
