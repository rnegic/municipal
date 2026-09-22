import { ErrorIllustration } from '@/shared/assets/illustrations'
import { commonTexts } from '@/shared/config/texts'
import { describeApiError } from '@/shared/lib/api-error'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'

export interface ApiErrorStateProps {
  error: unknown
  onRetry: () => void
  className?: string
}

export const ApiErrorState = ({ error, onRetry, className }: ApiErrorStateProps) => {
  const { title, description } = describeApiError(error)

  return (
    <EmptyState
      className={className}
      tone="danger"
      illustration={<ErrorIllustration />}
      title={title}
      description={description}
      action={
        <Button tone="secondary" size="medium" onClick={onRetry}>
          {commonTexts.actions.tryAgain}
        </Button>
      }
    />
  )
}
