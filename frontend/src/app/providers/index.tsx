import { compose } from './compose'
import { WithErrorBoundary } from './with-error-boundary'
import { WithMaxUI } from './with-max-ui'
import { WithQuery } from './with-query'
import { WithRouter } from './with-router'
import { WithTheme } from './with-theme'

export const withProviders = compose(WithTheme, WithMaxUI, WithErrorBoundary, WithQuery, WithRouter)
