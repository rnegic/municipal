import { Typography } from '@maxhub/max-ui'
import type { ReactNode } from 'react'

import { useUkSession } from '@/entities/session'
import { IconShield } from '@/shared/assets/icons'
import { ukAuthTexts as texts } from '../../config/texts'
import s from './UkSessionBadge.module.scss'

export interface UkSessionBadgeProps {
  actions?: ReactNode
}

export const UkSessionBadge = ({ actions }: UkSessionBadgeProps) => {
  const session = useUkSession()

  if (!session) {
    return null
  }

  const { organization, user, authMethod } = session

  return (
    <div className={s.root}>
      <div className={s.identity}>
        <span className={s.method}>
          <IconShield size={14} />
          {authMethod === 'esia' ? texts.session.methodEsia : texts.session.methodMock}
        </span>
        <Typography.Text variant="body-strong">{organization.name}</Typography.Text>
        <Typography.Text variant="note" color="tertiary">
          {texts.session.inn} {organization.inn}
          {organization.licenseNumber ? ` · ${texts.session.license} ${organization.licenseNumber}` : ''}
        </Typography.Text>
        <Typography.Text variant="note" color="tertiary">
          {user.fullName}
          {user.position ? ` · ${user.position}` : ''}
        </Typography.Text>
      </div>
      {actions ? <div className={s.actions}>{actions}</div> : null}
    </div>
  )
}
