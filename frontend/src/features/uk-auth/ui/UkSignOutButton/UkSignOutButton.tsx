import { IconButton } from '@maxhub/max-ui'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { IconSignOut } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { useUkSignOut } from '../../api'
import { ukAuthTexts as texts } from '../../config/texts'

export const UkSignOutButton = () => {
  const navigate = useNavigate()
  const signOut = useUkSignOut()
  const [isPending, setIsPending] = useState(false)

  return (
    <IconButton
      size="small"
      variant="secondary"
      aria-label={texts.session.signOut}
      loading={isPending}
      onClick={() => {
        setIsPending(true)
        void signOut().finally(() => {
          setIsPending(false)
          navigate(ROUTES.onboarding, { replace: true })
        })
      }}
    >
      <IconSignOut />
    </IconButton>
  )
}
