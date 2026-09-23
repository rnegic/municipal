import { useEffect, useId, useRef, type ReactNode } from 'react'

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
  const dialogRef = useRef<HTMLDialogElement>(null)
  const titleId = useId()
  const platform = usePlatform()

  useEffect(() => {
    const dialog = dialogRef.current

    if (!open || !dialog) {
      return undefined
    }

    const restoreOverflow = document.body.style.overflow

    document.body.style.overflow = 'hidden'
    dialog.showModal()

    return () => {
      document.body.style.overflow = restoreOverflow
      dialog.close()
    }
  }, [open])

  if (!open) {
    return null
  }

  return (
    <dialog
      ref={dialogRef}
      className={s.root}
      aria-labelledby={titleId}
      onCancel={(event) => {
        event.preventDefault()
        onClose()
      }}
      onClick={(event) => {
        if (event.target === dialogRef.current) {
          onClose()
        }
      }}
    >
      <section className={s.sheet}>
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
    </dialog>
  )
}