import type { ReactNode } from 'react'

import { cn } from '@/shared/lib/cn'
import s from './PageLayout.module.scss'

export type PageLayoutWidth = 'normal' | 'wide'

export interface PageLayoutProps {
  hero?: ReactNode
  header?: ReactNode
  floatingAction?: ReactNode
  children: ReactNode
  className?: string
  width?: PageLayoutWidth
}

export const PageLayout = ({
  hero,
  header,
  floatingAction,
  children,
  className,
  width = 'normal',
}: PageLayoutProps) => (
  <div className={cn(s.root, className)}>
    {hero}
    <div className={cn(s.content, hero ? s.contentWithHero : undefined, width === 'wide' && s.widthWide)}>
      {header ? <div className={s.header}>{header}</div> : null}
      <div className={s.body}>{children}</div>
      {floatingAction ? <div className={s.floatingActionSpacer} aria-hidden="true" /> : null}
    </div>
    {floatingAction ? (
      <div className={s.floatingAction}>
        <div className={cn(s.floatingActionInner, width === 'wide' && s.widthWide)}>{floatingAction}</div>
      </div>
    ) : null}
  </div>
)
