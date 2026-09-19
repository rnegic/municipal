import { Spinner, Typography } from '@maxhub/max-ui'

import { commonTexts } from '@/shared/config/texts'
import { cn } from '@/shared/lib/cn'
import s from './LoadingState.module.scss'

export interface LoadingStateProps {
  label?: string
  className?: string
}

export const LoadingState = ({
  label = commonTexts.states.loading,
  className,
}: LoadingStateProps) => (
  <div className={cn(s.root, className)} role="status">
    <Spinner size={24} appearance="themed" />
    <Typography.Text variant="description" color="secondary">
      {label}
    </Typography.Text>
  </div>
)
