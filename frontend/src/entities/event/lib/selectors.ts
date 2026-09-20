import type { Event } from '../model/schema'

const ACTIVE_STATUSES: readonly Event['status'][] = ['planned', 'in_progress']

const toTime = (event: Event): number => new Date(event.scheduledFrom).getTime()

export const selectActiveEvents = (events: readonly Event[]): Event[] =>
  events
    .filter((event) => ACTIVE_STATUSES.includes(event.status))
    .sort((a, b) => toTime(a) - toTime(b))
