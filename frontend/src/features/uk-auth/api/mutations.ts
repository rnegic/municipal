import { useMutation, useQueryClient } from '@tanstack/react-query'

import { clearUkSession, saveUkSession, ukSessionSchema } from '@/entities/session'
import { apiRequest } from '@/shared/api/client'
import { normalizeInn } from '../lib/inn'

export interface EsiaLoginInput {
  inn: string
  password: string
}

export const esiaLogin = (input: EsiaLoginInput) =>
  apiRequest(
    '/auth/esia-mock',
    { method: 'POST', body: { inn: normalizeInn(input.inn), password: input.password } },
    ukSessionSchema,
  )

export const useEsiaLoginMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: esiaLogin,
    onSuccess: async (session) => {
      saveUkSession(session)
      await queryClient.resetQueries()
    },
  })
}

export const useUkSignOut = () => {
  const queryClient = useQueryClient()

  return async () => {
    clearUkSession()
    await queryClient.resetQueries()
  }
}
