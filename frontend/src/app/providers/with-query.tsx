import { useState } from 'react'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { createProvider, type ProviderWrapperProps } from './create-provider'

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: false,
      },
    },
  })

const QueryProvider = ({ children }: ProviderWrapperProps) => {
  const [queryClient] = useState(createQueryClient)

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}

export const WithQuery = createProvider(QueryProvider)
