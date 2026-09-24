import { useMutation, useQueryClient } from '@tanstack/react-query'

import { clearBoundHouse } from '@/entities/house'
import { apiRequest } from '@/shared/api/client'

const UNBIND_HOUSE_PATH = '/houses/bind'

export const unbindHouse = () => apiRequest<void>(UNBIND_HOUSE_PATH, { method: 'DELETE' })

export const useResidentSignOut = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: unbindHouse,
    onSuccess: async () => {
      clearBoundHouse()
      await queryClient.resetQueries()
    },
  })
}
