export {
  INCIDENT_AUTHORITY_VALUES,
  INCIDENT_CATEGORY_VALUES,
  INCIDENT_DESCRIPTION_MAX_LENGTH,
  INCIDENT_DESCRIPTION_MIN_LENGTH,
  INCIDENT_ENTRANCE_MAX_LENGTH,
  INCIDENT_FLOOR_ZONE_MAX_LENGTH,
  INCIDENT_PHOTOS_MAX_COUNT,
  INCIDENT_PHOTO_MAX_BYTES,
  INCIDENT_PHOTO_MIME_TYPES,
  UK_AUTHORITY,
} from './config/domain'
export { incidentTexts } from './config/texts'
export {
  REQUESTS_PAGE_SIZE,
  analyzeIncident,
  createIncident,
  incidentKeys,
  uploadIncidentPhotos,
  useActiveIncidentsQuery,
  useAnalyzeIncidentMutation,
  useConfirmIncidentMutation,
  useCreateIncidentMutation,
  useIncidentQuery,
  useJoinIncidentMutation,
  useMyRequestsQuery,
  useUkQueueQuery,
  useSetIncidentStatusMutation,
  useUploadIncidentPhotoMutation,
} from './api'
export {
  incidentAuthorityOptions,
  incidentCategoryOptions,
  type IncidentDomainOption,
} from './lib/options'
export { isPhotoRequiredForCategory, isUkAuthority } from './lib/routing'
export type {
  AnalyzeIncidentInput,
  AnalyzeIncidentResult,
  ConfirmResponse,
  CreateIncidentInput,
  Incident,
  IncidentAuthority,
  IncidentCategory,
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
