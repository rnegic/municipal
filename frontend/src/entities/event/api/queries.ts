import { skipToken, useQuery } from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import { eventListResponseSchema } from '../model/schema'
import { eventKeys } from './keys'

const EVENTS_STALE_TIME = 30_000
const EVENTS_REFETCH_INTERVAL = 60_000

export const fetchEvents = (houseId: string) =>
  apiRequest('/events', { method: 'GET', query: { houseId } }, eventListResponseSchema)

export const useEventsQuery = (houseId: string | undefined) =>
  useQuery({
    queryKey: eventKeys.list(houseId ?? ''),
    queryFn: houseId ? () => fetchEvents(houseId) : skipToken,
    staleTime: EVENTS_STALE_TIME,
    refetchInterval: EVENTS_REFETCH_INTERVAL,
  })
