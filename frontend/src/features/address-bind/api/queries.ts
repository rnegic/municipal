import { skipToken, useQuery, type QueryFunctionContext } from '@tanstack/react-query'

import { fetchAddressSuggestions } from '@/shared/lib/dadata'
import { useDebounce } from '@/shared/lib/hooks'

export const ADDRESS_SUGGEST_DEBOUNCE_MS = 350
export const ADDRESS_SUGGEST_MIN_LENGTH = 3

const SUGGEST_COUNT = 5
const SUGGEST_STALE_TIME = 5 * 60 * 1000

export const addressSuggestKeys = {
  all: ['addressSuggest'] as const,
  byQuery: (query: string) => [...addressSuggestKeys.all, query] as const,
}

export const useAddressSuggestionsQuery = (value: string) => {
  const query = value.trim()
  const debouncedQuery = useDebounce(query, ADDRESS_SUGGEST_DEBOUNCE_MS)
  const isQueryReady = debouncedQuery.length >= ADDRESS_SUGGEST_MIN_LENGTH

  const { data, isFetching } = useQuery({
    queryKey: addressSuggestKeys.byQuery(debouncedQuery),
    queryFn: isQueryReady
      ? ({ signal }: QueryFunctionContext) =>
          fetchAddressSuggestions({ query: debouncedQuery, count: SUGGEST_COUNT, signal })
      : skipToken,
    staleTime: SUGGEST_STALE_TIME,
    retry: false,
  })

  return {
    suggestions: data?.suggestions ?? [],
    isFetching: isQueryReady && isFetching,
    isSettled: isQueryReady && debouncedQuery === query && !isFetching && data !== undefined,
  }
}