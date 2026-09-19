import { cn } from '@/shared/lib/cn'
import s from './StatusDot.module.scss'

export type StatusTone = 'neutral' | 'success' | 'danger' | 'progress'

export type StatusDotSize = 'small' | 'medium'

export interface StatusDotProps {
  tone?: StatusTone
  size?: StatusDotSize
  pulse?: boolean
  className?: string
}

const TONE_CLASS: Record<StatusTone, string | undefined> = {
  neutral: undefined,
  success: s.toneSuccess,
  danger: s.toneDanger,
  progress: s.toneProgress,
}

const SIZE_CLASS: Record<StatusDotSize, string> = {
  small: s.sizeSmall,
  medium: s.sizeMedium,
}

export const StatusDot = ({
  tone = 'neutral',
  size = 'small',
  pulse = false,
  className,
}: StatusDotProps) => (
  <span
    className={cn(s.root, TONE_CLASS[tone], SIZE_CLASS[size], pulse && s.pulse, className)}
    aria-hidden="true"
  />
)
