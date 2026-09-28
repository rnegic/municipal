import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import { apiKeyListSchema, createdApiKeySchema } from '../model/schema'

const apiKeyKeys = {
  all: ['uk-api-keys'] as const,
}

export const fetchApiKeys = () =>
  apiRequest('/uk/api-keys', { method: 'GET', auth: 'uk' }, apiKeyListSchema)

export const createApiKey = (name: string) =>
  apiRequest('/uk/api-keys', { method: 'POST', body: { name }, auth: 'uk' }, createdApiKeySchema)

export const revokeApiKey = (id: string) =>
  apiRequest<void>(`/uk/api-keys/${id}`, { method: 'DELETE', auth: 'uk' })

export const useApiKeysQuery = (enabled: boolean) =>
  useQuery({
    queryKey: apiKeyKeys.all,
    queryFn: fetchApiKeys,
    enabled,
  })

export const useCreateApiKeyMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createApiKey,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: apiKeyKeys.all }),
  })
}

export const useRevokeApiKeyMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: revokeApiKey,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: apiKeyKeys.all }),
  })
}
