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
  IncidentStatusBadge,
  type IncidentStatusBadgeProps,
} from './ui/IncidentStatusBadge'
