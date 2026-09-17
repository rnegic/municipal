import { compose } from './compose'
import { WithErrorBoundary } from './with-error-boundary'
import { WithMaxUI } from './with-max-ui'
import { WithQuery } from './with-query'
import { WithRouter } from './with-router'

export const withProviders = compose(WithMaxUI, WithErrorBoundary, WithQuery, WithRouter)
