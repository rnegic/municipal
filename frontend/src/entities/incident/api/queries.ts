import { skipToken, useQuery } from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import {
  incidentListResponseSchema,
  incidentSchema,
  paginatedUkQueueSchema,
  paginatedResidentRequestsSchema,
} from '../model/schema'
import { incidentKeys } from './keys'

export const REQUESTS_PAGE_SIZE = 20

const ACTIVE_INCIDENTS_STALE_TIME = 15_000
const ACTIVE_INCIDENTS_REFETCH_INTERVAL = 30_000
const REQUESTS_STALE_TIME = 30_000
const REQUESTS_REFETCH_INTERVAL = 10_000
const INCIDENT_STALE_TIME = 15_000
const INCIDENT_REFETCH_INTERVAL = 30_000
const UK_QUEUE_STALE_TIME = 10_000
const UK_QUEUE_REFETCH_INTERVAL = 20_000

export const fetchActiveIncidents = (houseId: string) =>
  apiRequest(
    `/houses/${houseId}/incidents`,
    { method: 'GET', query: { status: 'active' } },
    incidentListResponseSchema,
  )

export const fetchMyRequests = (
  houseId: string,
  offset = 0,
  limit = REQUESTS_PAGE_SIZE,
) =>
  apiRequest(
    `/houses/${houseId}/requests`,
    { method: 'GET', query: { offset, limit } },
    paginatedResidentRequestsSchema,
  )

export const fetchIncident = (incidentId: string) =>
  apiRequest(`/incidents/${incidentId}`, { method: 'GET' }, incidentSchema)

export const fetchUkQueue = (offset = 0, limit = 100) =>
  apiRequest(
    '/uk/queue',
    { method: 'GET', query: { offset, limit } },
    paginatedUkQueueSchema,
  )

export const useActiveIncidentsQuery = (houseId: string | undefined) =>
  useQuery({
    queryKey: incidentKeys.list(houseId ?? ''),
    queryFn: houseId ? () => fetchActiveIncidents(houseId) : skipToken,
    staleTime: ACTIVE_INCIDENTS_STALE_TIME,
    refetchInterval: ACTIVE_INCIDENTS_REFETCH_INTERVAL,
  })

export const useMyRequestsQuery = (
  houseId: string | undefined,
  offset = 0,
  limit = REQUESTS_PAGE_SIZE,
) =>
  useQuery({
    queryKey: incidentKeys.requests(houseId ?? '', offset, limit),
    queryFn: houseId ? () => fetchMyRequests(houseId, offset, limit) : skipToken,
    staleTime: REQUESTS_STALE_TIME,
    refetchInterval: REQUESTS_REFETCH_INTERVAL,
  })

export const useIncidentQuery = (incidentId: string | undefined) =>
  useQuery({
    queryKey: incidentKeys.byId(incidentId ?? ''),
    queryFn: incidentId ? () => fetchIncident(incidentId) : skipToken,
    staleTime: INCIDENT_STALE_TIME,
    refetchInterval: INCIDENT_REFETCH_INTERVAL,
  })

export const useUkQueueQuery = (offset = 0, limit = 100) =>
  useQuery({
    queryKey: incidentKeys.ukQueue(offset, limit),
    queryFn: () => fetchUkQueue(offset, limit),
    staleTime: UK_QUEUE_STALE_TIME,
    refetchInterval: UK_QUEUE_REFETCH_INTERVAL,
  })
