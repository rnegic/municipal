import type { ZodType } from 'zod'

import { env } from '@/shared/config/env'
import { getInitData } from '@/shared/lib/max'

export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

export type QueryValue = string | number | boolean | undefined

export interface ApiErrorPayload {
  code: string
  message: string
  details?: unknown
}

export interface ApiRequestOptions<TBody = never> {
  method?: HttpMethod
  body?: TBody
  query?: Record<string, QueryValue>
  headers?: Record<string, string>
  signal?: AbortSignal
  withAuth?: boolean
}

export class ApiError extends Error {
  readonly status: number
  readonly payload: ApiErrorPayload | null

  constructor(status: number, payload: ApiErrorPayload | null) {
    super(payload?.message ?? `HTTP ${status}`)
    this.name = 'ApiError'
    this.status = status
    this.payload = payload
  }
}

const INIT_DATA_HEADER = 'Authorization'

const buildUrl = (path: string, query?: Record<string, QueryValue>): string => {
  const url = new URL(`${env.apiBaseUrl}${path}`, window.location.origin)

  for (const [key, value] of Object.entries(query ?? {})) {
    if (value !== undefined) {
      url.searchParams.set(key, String(value))
    }
  }

  return url.toString()
}

const parsePayload = async (response: Response): Promise<ApiErrorPayload | null> => {
  try {
    return (await response.json()) as ApiErrorPayload
  } catch {
    return null
  }
}

/**
 * Единственный HTTP-клиент приложения
 */
export const apiRequest = async <TResponse, TBody = never>(
  path: string,
  options: ApiRequestOptions<TBody> = {},
  schema?: ZodType<TResponse>,
): Promise<TResponse> => {
  const { method = 'GET', body, query, headers, signal, withAuth = true } = options
  const initData = withAuth ? getInitData() : ''

  const response = await fetch(buildUrl(path, query), {
    method,
    signal,
    headers: {
      Accept: 'application/json',
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      ...(initData ? { [INIT_DATA_HEADER]: `tma ${initData}` } : {}),
      ...headers,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  if (!response.ok) {
    throw new ApiError(response.status, await parsePayload(response))
  }

  if (response.status === 204) {
    return undefined as TResponse
  }

  const payload: unknown = await response.json()

  return schema ? schema.parse(payload) : (payload as TResponse)
}
