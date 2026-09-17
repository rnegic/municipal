import type { ComponentType } from 'react'

export type Provider = (Component: ComponentType) => ComponentType

export const compose =
  (...providers: readonly Provider[]) =>
  (Component: ComponentType): ComponentType =>
    providers.reduceRight<ComponentType>((Wrapped, provider) => provider(Wrapped), Component)
