import type { ComponentProps } from 'react'

import { cn } from '@/shared/lib/cn'
import s from './Card.module.scss'

export type CardTone = 'default' | 'success' | 'danger' | 'progress'

export type CardPadding = 'default' | 'compact'

export interface CardProps extends ComponentProps<'div'> {
  tone?: CardTone
  padding?: CardPadding
}

const TONE_CLASS: Record<CardTone, string | undefined> = {
  default: undefined,
  success: s.toneSuccess,
  danger: s.toneDanger,
  progress: s.toneProgress,
}

const PADDING_CLASS: Record<CardPadding, string> = {
  default: s.paddingDefault,
  compact: s.paddingCompact,
}

export const Card = ({ tone = 'default', padding = 'default', className, ...rest }: CardProps) => (
  <div className={cn(s.root, TONE_CLASS[tone], PADDING_CLASS[padding], className)} {...rest} />
)
