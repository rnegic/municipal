import { ApiError } from '@/shared/api/client'
import { commonTexts } from '@/shared/config/texts'

export interface ApiErrorDescription {
  title: string
  description: string
}

const { errors } = commonTexts

const withServerMessage = (
  title: string,
  fallbackDescription: string,
  error: ApiError,
): ApiErrorDescription => ({
  title,
  description: error.payload?.message || fallbackDescription,
})

export const describeApiError = (error: unknown): ApiErrorDescription => {
  if (!(error instanceof ApiError)) {
    return { title: errors.networkTitle, description: errors.networkDescription }
  }

  switch (error.status) {
    case 400:
      return withServerMessage(errors.validationTitle, errors.validationDescription, error)
    case 401:
      return { title: errors.unauthorizedTitle, description: errors.unauthorizedDescription }
    case 403:
      return { title: errors.forbiddenTitle, description: errors.forbiddenDescription }
    case 404:
      return { title: errors.notFoundTitle, description: errors.notFoundDescription }
    case 422:
      return withServerMessage(errors.businessTitle, errors.businessDescription, error)
    case 429:
      return withServerMessage(errors.rateLimitTitle, errors.rateLimitDescription, error)
    default:
      return { title: errors.unexpectedTitle, description: errors.unexpectedDescription }
  }
}
