export { incidentTexts } from './config/texts'
export {
  REQUESTS_PAGE_SIZE,
  incidentKeys,
  useActiveIncidentsQuery,
  useConfirmIncidentMutation,
  useCreateIncidentMutation,
  useIncidentQuery,
  useJoinIncidentMutation,
  useMyRequestsQuery,
  useUkQueueQuery,
  useSetIncidentStatusMutation,
  useUploadIncidentPhotoMutation,
} from './api'
export type {
  ConfirmResponse,
  CreateIncidentInput,
  Incident,
  IncidentPhoto,
  IncidentListResponse,
  IncidentSeverity,
  IncidentStatus,
  JoinResponse,
  PaginatedResidentRequests,
  ResidentRequest,
  PaginatedUkQueue,
  UkQueueItem,
} from './model/types'
export {
  IncidentPhotos,
} from './ui/IncidentPhotos'
export { IncidentStatusBadge, type IncidentStatusBadgeProps } from './ui/IncidentStatusBadge'
