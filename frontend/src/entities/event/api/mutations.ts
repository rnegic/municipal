import { useMutation, useQueryClient } from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import { createEventRequestSchema, eventSchema, type CreateEventInput } from '../model/schema'
import { eventKeys } from './keys'

export const createEvent = (input: CreateEventInput) =>
  apiRequest(
    '/uk/events',
    { method: 'POST', body: createEventRequestSchema.parse(input) },
    eventSchema,
  )

export const useCreateEventMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createEvent,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: eventKeys.all }),
  })
}
