import type { ReactNode } from 'react'

import { Typography } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import s from './Section.module.scss'

export type SectionTone = 'default' | 'inverse'

export interface SectionProps {
  title: string
  action?: ReactNode
  tone?: SectionTone
  children: ReactNode
  className?: string
}

const TONE_CLASS: Record<SectionTone, string | undefined> = {
  default: undefined,
  inverse: s.toneInverse,
}

export const Section = ({ title, action, tone = 'default', children, className }: SectionProps) => (
  <section className={cn(s.root, TONE_CLASS[tone], className)}>
    <header className={s.header}>
      <Typography.Title variant="large-strong">{title}</Typography.Title>
      {action ? <div className={s.action}>{action}</div> : null}
    </header>
    <div className={s.body}>{children}</div>
  </section>
)
