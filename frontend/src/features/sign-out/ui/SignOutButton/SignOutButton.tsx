import { IconButton } from '@maxhub/max-ui'
import { useNavigate } from 'react-router-dom'

import { IconSignOut } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { useResidentSignOut } from '../../api'

export const SignOutButton = () => {
  const navigate = useNavigate()
  const signOut = useResidentSignOut()

  return (
    <IconButton
      size="small"
      variant="secondary"
      aria-label={commonTexts.actions.signOut}
      loading={signOut.isPending}
      onClick={() => {
        signOut.mutate(undefined, {
          onSuccess: () => navigate(ROUTES.onboarding, { replace: true }),
        })
      }}
    >
      <IconSignOut />
    </IconButton>
  )
}
