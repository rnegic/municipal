import type { IncidentStatus } from '../model/schema'

const CLOSED_STATUSES: ReadonlySet<IncidentStatus> = new Set<IncidentStatus>(['done', 'false_alarm'])

export const isIncidentClosed = (status: IncidentStatus): boolean => CLOSED_STATUSES.has(status)

export const isIncidentFalseAlarm = (status: IncidentStatus): boolean => status === 'false_alarm'
