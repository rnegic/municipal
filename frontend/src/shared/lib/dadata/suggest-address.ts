import { z } from 'zod'

import { apiRequest } from '@/shared/api/client'

export const ADDRESS_SUGGEST_PATH = '/houses/suggest'

const addressSuggestionSchema = z.object({
  value: z.string().min(1),
  houseFiasId: z.string().nullish(),
})

const addressSuggestResponseSchema = z.object({
  suggestions: z.array(addressSuggestionSchema),
})

export type AddressSuggestion = z.infer<typeof addressSuggestionSchema>

export interface AddressSuggestParams {
  query: string
  count: number
  signal?: AbortSignal
}

export const fetchAddressSuggestions = ({ query, count, signal }: AddressSuggestParams) =>
  apiRequest(
    ADDRESS_SUGGEST_PATH,
    { method: 'GET', query: { query, count }, signal },
    addressSuggestResponseSchema,
  )

export interface AddressByCoordsParams {
  lat: number
  lon: number
  count: number
  signal?: AbortSignal
}

/** Ближайший адрес с домом по координатам (обратный геокодинг через бэкенд). */
export const fetchAddressByCoords = ({ lat, lon, count, signal }: AddressByCoordsParams) =>
  apiRequest(
    ADDRESS_SUGGEST_PATH,
    { method: 'GET', query: { lat, lon, count }, signal },
    addressSuggestResponseSchema,
  )