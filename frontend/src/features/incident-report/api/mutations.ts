import { useMutation, useQueryClient } from '@tanstack/react-query'

import {
  createIncident,
  incidentKeys,
  uploadIncidentPhotos,
  type IncidentCategory,
} from '@/entities/incident'

export interface SubmitIncidentReportInput {
  description: string
  category: IncidentCategory
  entrance?: string
  floorZone?: string
  photos: File[]
}

export const submitIncidentReport = async ({ photos, ...incident }: SubmitIncidentReportInput) => {
  const photoUrls = photos.length > 0 ? await uploadIncidentPhotos(photos) : []

  return createIncident({ ...incident, photoUrls })
}

export const useSubmitIncidentReportMutation = () => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: submitIncidentReport,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: incidentKeys.all }),
  })
}
