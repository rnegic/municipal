export interface ApiErrorResponse {
  code: string
  message: string
  details?: unknown
}

export interface PaginatedResponse<TItem> {
  items: TItem[]
  total: number
  offset: number
  limit: number
}

export interface IncidentJoinResponse {
  incidentId: string
  affectedCount: number
  joined: boolean
}
