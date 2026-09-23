import { Typography } from '@maxhub/max-ui'
import { useNavigate } from 'react-router-dom'

import { useUkSession } from '@/entities/session'
import { IconShield } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { Button } from '@/shared/ui/button'
import { useUkSignOut } from '../../api'
import { ukAuthTexts as texts } from '../../config/texts'
import s from './UkSessionBadge.module.scss'

export const UkSessionBadge = () => {
  const session = useUkSession()
  const signOut = useUkSignOut()
  const navigate = useNavigate()

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
      <Button
        type="button"
        size="small"
        tone="secondary"
        onClick={() => {
          void signOut().then(() => navigate(ROUTES.onboarding, { replace: true }))
        }}
      >
        {texts.session.signOut}
      </Button>
    </div>
  )
}
