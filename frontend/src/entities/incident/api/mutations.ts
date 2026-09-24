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
  type AnalyzeIncidentInput,
  type CreateIncidentInput,
  type Incident,
  type IncidentListResponse,
  type PaginatedUkQueue,
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
      await queryClient.cancelQueries({ queryKey: incidentKeys.ukQueue(0, 100) })
      const queryKey = incidentKeys.ukQueue(0, 100)
      const previous = queryClient.getQueryData<PaginatedUkQueue>(queryKey)

      queryClient.setQueryData<PaginatedUkQueue>(queryKey, (data) =>
        data
          ? { ...data, items: data.items.map((item) => (item.id === incidentId ? { ...item, status } : item)) }
          : data,
      )

      return { previous, queryKey }
    },
    onError: (_error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(context.queryKey, context.previous)
      }
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}
