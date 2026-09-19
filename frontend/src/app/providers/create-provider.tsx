import type { ComponentType, ReactNode } from 'react'

import type { Provider } from './compose'

export interface ProviderWrapperProps {
  children: ReactNode
}

export type ProviderWrapper = ComponentType<ProviderWrapperProps>

export const createProvider =
  (Wrapper: ProviderWrapper): Provider =>
  (Component: ComponentType) => {
    const WithWrapper = () => (
      <Wrapper>
        <Component />
      </Wrapper>
    )

    WithWrapper.displayName = `with(${Component.displayName ?? Component.name ?? 'Component'})`

    return WithWrapper
  }
