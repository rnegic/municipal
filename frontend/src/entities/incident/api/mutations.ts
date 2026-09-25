import {
  useMutation,
  useQueryClient,
  type QueryClient,
  type QueryKey,
} from '@tanstack/react-query'

import { apiRequest } from '@/shared/api/client'
import {
  analyzeIncidentRequestSchema,
  analyzeIncidentResponseSchema,
  confirmResponseSchema,
  createIncidentRequestSchema,
  incidentSchema,
  incidentPhotoSchema,
  joinResponseSchema,
  mergeIncidentsRequestSchema,
  mergeIncidentsResponseSchema,
  type AnalyzeIncidentInput,
  type CreateIncidentInput,
  type Incident,
  type IncidentListResponse,
  type MergeIncidentsInput,
  type PaginatedUkQueue,
  type UkQueueItem,
} from '../model/schema'
import { incidentKeys } from './keys'

interface IncidentMutationContext {
  previous: Array<[QueryKey, unknown]>
}

const ANALYZE_TIMEOUT_MS = 20_000

export const joinIncident = (incidentId: string) =>
  apiRequest(`/incidents/${incidentId}/join`, { method: 'POST' }, joinResponseSchema)

export const confirmIncident = (incidentId: string) =>
  apiRequest(`/incidents/${incidentId}/confirm`, { method: 'POST' }, confirmResponseSchema)

export const analyzeIncident = (input: AnalyzeIncidentInput, signal?: AbortSignal) =>
  apiRequest(
    '/incidents/analyze',
    { method: 'POST', body: analyzeIncidentRequestSchema.parse(input), signal },
    analyzeIncidentResponseSchema,
  )

export const createIncident = (input: CreateIncidentInput) =>
  apiRequest(
    '/incidents',
    { method: 'POST', body: createIncidentRequestSchema.parse(input) },
    incidentSchema,
  )

export const uploadIncidentPhoto = (incidentId: string, file: File) => {
  const body = new FormData()
  body.append('photo', file)

  return apiRequest(
    `/incidents/${incidentId}/photos`,
    { method: 'POST', body },
    incidentPhotoSchema,
  )
}

export const uploadIncidentPhotoFile = (file: File) => {
  const body = new FormData()
  body.append('photo', file)

  return apiRequest('/incidents/photos', { method: 'POST', body }, incidentPhotoSchema)
}

export const uploadIncidentPhotos = async (files: File[]): Promise<string[]> => {
  const uploaded = await Promise.all(files.map(uploadIncidentPhotoFile))

  return uploaded.map((photo) => photo.url)
}

export const setIncidentStatus = (incidentId: string, status: Incident['status']) =>
  apiRequest(
    `/incidents/${incidentId}/status`,
    { method: 'PATCH', body: { status }, auth: 'uk' },
    incidentSchema,
  )

export const mergeIncidents = (input: MergeIncidentsInput) =>
  apiRequest(
    '/uk/incidents/merge',
    { method: 'POST', body: mergeIncidentsRequestSchema.parse(input), auth: 'uk' },
    mergeIncidentsResponseSchema,
  )

const patchUkQueueCaches = (
  queryClient: QueryClient,
  patch: (items: readonly UkQueueItem[]) => UkQueueItem[],
): void => {
  queryClient.setQueriesData<PaginatedUkQueue>({ queryKey: incidentKeys.ukQueues() }, (data) => {
    if (!data) {
      return data
    }

    const items = patch(data.items)

    return { ...data, items, total: data.total - (data.items.length - items.length) }
  })
}

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

export const useAnalyzeIncidentMutation = () =>
  useMutation({
    mutationFn: (input: AnalyzeIncidentInput) =>
      analyzeIncident(input, AbortSignal.timeout(ANALYZE_TIMEOUT_MS)),
    retry: false,
  })

export const useCreateIncidentMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createIncident,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}

export const useUploadIncidentPhotoMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ incidentId, file }: { incidentId: string; file: File }) =>
      uploadIncidentPhoto(incidentId, file),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}

export const useSetIncidentStatusMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ incidentId, status }: { incidentId: string; status: Incident['status'] }) =>
      setIncidentStatus(incidentId, status),
    onMutate: async ({ incidentId, status }) => {
      const context = await snapshotIncidentCaches(queryClient)

      patchUkQueueCaches(queryClient, (items) =>
        items.map((item) => (item.id === incidentId ? { ...item, status } : item)),
      )

      return context
    },
    onError: (_error, _variables, context) => {
      restoreIncidentCaches(queryClient, context)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}

export const useMergeIncidentsMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: mergeIncidents,
    onMutate: async ({ targetIncidentId, sourceIncidentIds }) => {
      const context = await snapshotIncidentCaches(queryClient)
      const sources = new Set(sourceIncidentIds)

      patchUkQueueCaches(queryClient, (items) => {
        const merged = items.filter((item) => sources.has(item.id))
        const affected = merged.reduce((sum, item) => sum + item.affectedCount, 0)
        const duplicates = merged.reduce((sum, item) => sum + item.mergedCount, 0)

        return items
          .filter((item) => !sources.has(item.id))
          .map((item) =>
            item.id === targetIncidentId
              ? {
                  ...item,
                  affectedCount: item.affectedCount + affected,
                  mergedCount: item.mergedCount + merged.length + duplicates,
                }
              : item,
          )
      })

      return context
    },
    onError: (_error, _variables, context) => {
      restoreIncidentCaches(queryClient, context)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}
