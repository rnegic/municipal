import { useMutation, useQueryClient } from '@tanstack/react-query'

import { houseSchema } from '@/entities/house'
import { userKeys } from '@/entities/user'
import { apiRequest } from '@/shared/api/client'

export const bindHouse = (address: string) =>
  apiRequest('/houses/bind', { method: 'POST', body: { address } }, houseSchema)

export const useBindHouseMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: bindHouse,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: userKeys.me() }),
  })
}
