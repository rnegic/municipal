import {
  useMutation,
  useQueryClient,
  type QueryClient,
  type QueryKey,
} from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import {
  confirmResponseSchema,
  createIncidentRequestSchema,
  incidentSchema,
  joinResponseSchema,
  type CreateIncidentInput,
  type Incident,
  type IncidentListResponse,
} from '../model/schema'
import { incidentKeys } from './keys'

interface IncidentMutationContext {
  previous: Array<[QueryKey, unknown]>
}

export const joinIncident = (incidentId: string) =>
  apiRequest(`/incidents/${incidentId}/join`, { method: 'POST' }, joinResponseSchema)

export const confirmIncident = (incidentId: string) =>
  apiRequest(`/incidents/${incidentId}/confirm`, { method: 'POST' }, confirmResponseSchema)

export const createIncident = (input: CreateIncidentInput) =>
  apiRequest(
    '/incidents',
    { method: 'POST', body: createIncidentRequestSchema.parse(input) },
    incidentSchema,
  )

const patchIncidentCaches = (
  queryClient: QueryClient,
  incidentId: string,
  patch: (incident: Incident) => Incident,
): void => {
  queryClient.setQueriesData<IncidentListResponse>(
    { queryKey: incidentKeys.lists() },
    (data) =>
      data
        ? {
            ...data,
            items: data.items.map((item) => (item.id === incidentId ? patch(item) : item)),
          }
        : data,
  )

  queryClient.setQueriesData<Incident>({ queryKey: incidentKeys.details() }, (data) =>
    data && data.id === incidentId ? patch(data) : data,
  )
}

const snapshotIncidentCaches = async (
  queryClient: QueryClient,
): Promise<IncidentMutationContext> => {
  await queryClient.cancelQueries({ queryKey: incidentKeys.all })

  return { previous: queryClient.getQueriesData({ queryKey: incidentKeys.all }) }
}

const restoreIncidentCaches = (
  queryClient: QueryClient,
  context: IncidentMutationContext | undefined,
): void => {
  context?.previous.forEach(([queryKey, data]) => queryClient.setQueryData(queryKey, data))
}

export const useJoinIncidentMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: joinIncident,
    onMutate: async (incidentId) => {
      const context = await snapshotIncidentCaches(queryClient)

      patchIncidentCaches(queryClient, incidentId, (incident) =>
        incident.joinedByMe
          ? incident
          : { ...incident, joinedByMe: true, affectedCount: incident.affectedCount + 1 },
      )

      return context
    },
    onError: (_error, _incidentId, context) => {
      restoreIncidentCaches(queryClient, context)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}

export const useConfirmIncidentMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: confirmIncident,
    onMutate: async (incidentId) => {
      const context = await snapshotIncidentCaches(queryClient)

      patchIncidentCaches(queryClient, incidentId, (incident) =>
        incident.confirmedByMe ? incident : { ...incident, confirmedByMe: true },
      )

      return context
    },
    onError: (_error, _incidentId, context) => {
      restoreIncidentCaches(queryClient, context)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}

export const useCreateIncidentMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createIncident,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}
