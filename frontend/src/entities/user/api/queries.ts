import { useQuery } from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import { meResponseSchema } from '../model/schema'
import { userKeys } from './keys'

const ME_STALE_TIME = 5 * 60_000

export const fetchMe = () => apiRequest('/me', { method: 'GET' }, meResponseSchema)

export const useMeQuery = () =>
  useQuery({
    queryKey: userKeys.me(),
    queryFn: fetchMe,
    staleTime: ME_STALE_TIME,
  })

export const useCurrentHouseQuery = () => {
  const query = useMeQuery()

  return { ...query, data: query.data?.house ?? null }
}
