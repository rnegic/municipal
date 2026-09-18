import { useEffect, useId, type ReactNode } from 'react'

import {
  Icon16CloseIos,
  Icon24CloseAndroid,
  IconButton,
  Typography,
  usePlatform,
} from '@maxhub/max-ui'

import { commonTexts } from '@/shared/config/texts'
import s from './BottomSheet.module.scss'

export interface BottomSheetProps {
  open: boolean
  title: string
  description?: string
  children: ReactNode
  onClose: () => void
}

export const BottomSheet = ({ open, title, description, children, onClose }: BottomSheetProps) => {
  const titleId = useId()
  const platform = usePlatform()

  useEffect(() => {
    if (!open) {
      return undefined
    }

    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    const restoreOverflow = document.body.style.overflow

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', closeOnEscape)

    return () => {
      document.body.style.overflow = restoreOverflow
      window.removeEventListener('keydown', closeOnEscape)
    }
  }, [onClose, open])

  if (!open) {
    return null
  }

  return (
    <div className={s.root}>
      <div className={s.backdrop} aria-hidden="true" onClick={onClose} />
      <section className={s.sheet} role="dialog" aria-modal="true" aria-labelledby={titleId}>
        <div className={s.handle} aria-hidden="true" />
        <header className={s.header}>
          <div className={s.heading}>
            <Typography.Title id={titleId} variant="medium-strong">
              {title}
            </Typography.Title>
            {description ? (
              <Typography.Text variant="description" color="secondary">
                {description}
              </Typography.Text>
            ) : null}
          </div>
          <IconButton
            size="small"
            variant="secondary"
            aria-label={commonTexts.actions.close}
            onClick={onClose}
          >
            {platform === 'ios' ? <Icon16CloseIos /> : <Icon24CloseAndroid />}
          </IconButton>
        </header>
        {children}
      </section>
    </div>
  )
}
