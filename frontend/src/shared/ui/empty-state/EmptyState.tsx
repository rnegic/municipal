import type { ReactNode } from 'react'

import { Typography } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import s from './EmptyState.module.scss'

export type EmptyStateTone = 'neutral' | 'success' | 'danger' | 'inverse'

export interface EmptyStateProps {
  illustration?: ReactNode
  icon?: ReactNode
  title: string
  description?: string
  tone?: EmptyStateTone
  action?: ReactNode
  className?: string
}

const TONE_CLASS: Record<EmptyStateTone, string | undefined> = {
  neutral: undefined,
  success: s.toneSuccess,
  danger: s.toneDanger,
  inverse: s.toneInverse,
}

export const EmptyState = ({
  illustration,
  icon,
  title,
  description,
  tone = 'neutral',
  action,
  className,
}: EmptyStateProps) => (
  <div className={cn(s.root, TONE_CLASS[tone], className)}>
    {illustration ? <span className={s.illustration}>{illustration}</span> : null}
    {icon ? <span className={s.icon}>{icon}</span> : null}
    <div className={s.content}>
      <Typography.Text variant="body-strong">{title}</Typography.Text>
      {description ? (
        <Typography.Text variant="description" color="secondary">
          {description}
        </Typography.Text>
      ) : null}
    </div>
    {action ? <div className={s.action}>{action}</div> : null}
  </div>
)
