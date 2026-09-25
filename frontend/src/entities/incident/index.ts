export {
  INCIDENT_AUTHORITY_VALUES,
  INCIDENT_CATEGORY_VALUES,
  INCIDENT_DESCRIPTION_MAX_LENGTH,
  INCIDENT_DESCRIPTION_MIN_LENGTH,
  INCIDENT_ENTRANCE_MAX_LENGTH,
  INCIDENT_FLOOR_ZONE_MAX_LENGTH,
  INCIDENT_MERGE_MIN_COUNT,
  INCIDENT_PHOTOS_MAX_COUNT,
  INCIDENT_PHOTO_MAX_BYTES,
  INCIDENT_PHOTO_MIME_TYPES,
  INCIDENT_SUPPORTERS_PREVIEW_COUNT,
  INCIDENT_TITLE_MAX_LENGTH,
  INCIDENT_TITLE_MIN_LENGTH,
  UK_AUTHORITY,
  UK_QUEUE_PRIORITY_MIN_AFFECTED,
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
  useMergeIncidentsMutation,
  useMyRequestsQuery,
  useUkQueueQuery,
  useSetIncidentStatusMutation,
  useUploadIncidentPhotoMutation,
} from './api'
export {
  getMergeBlocker,
  getMergedAffectedCount,
  suggestMergeTargetId,
} from './lib/merge'
export {
  incidentAuthorityOptions,
  incidentCategoryOptions,
  type IncidentDomainOption,
} from './lib/options'
export { isPriorityQueueItem, sortUkQueueByPriority } from './lib/priority'
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
  IncidentSupporter,
  JoinResponse,
  MergeIncidentsInput,
  MergeIncidentsResult,
  PaginatedResidentRequests,
  ResidentRequest,
  PaginatedUkQueue,
  UkQueueItem,
} from './model/types'
export {
  IncidentPhotos,
} from './ui/IncidentPhotos'
export { IncidentStatusBadge, type IncidentStatusBadgeProps } from './ui/IncidentStatusBadge'
