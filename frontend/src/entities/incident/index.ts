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
} from './api'
export type {
  ConfirmResponse,
  CreateIncidentInput,
  Incident,
  IncidentListResponse,
  IncidentSeverity,
  IncidentStatus,
  JoinResponse,
  PaginatedResidentRequests,
  ResidentRequest,
} from './model/types'
export {
  IncidentStatusBadge,
  type IncidentStatusBadgeProps,
} from './ui/IncidentStatusBadge'
