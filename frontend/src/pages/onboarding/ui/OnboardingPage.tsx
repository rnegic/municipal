import { useState } from 'react'

import { Typography } from '@maxhub/max-ui'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { useCurrentHouseQuery } from '@/entities/user'
import { AddressBindSheet } from '@/features/address-bind'
import { UkAuthEntryLink, UkAuthSheet } from '@/features/uk-auth'
import { IconChevronLeft } from '@/shared/assets/icons'
import { OnboardingIllustration } from '@/shared/assets/illustrations'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { Button } from '@/shared/ui/button'
import { PageLayout } from '@/shared/ui/page-layout'
import { onboardingTexts } from '../config/texts'
import s from './OnboardingPage.module.scss'

interface OnboardingLocationState {
  ukAuth?: boolean
}

export const OnboardingPage = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const houseQuery = useCurrentHouseQuery()
  const hasAddress = houseQuery.data !== null
  const [isManualOpen, setIsManualOpen] = useState(false)
  const [isAutoDismissed, setIsAutoDismissed] = useState(false)
  const [isUkAuthOpen, setIsUkAuthOpen] = useState(
    Boolean((location.state as OnboardingLocationState | null)?.ukAuth),
  )

  const isAddressSheetOpen =
    !isUkAuthOpen && (isManualOpen || (houseQuery.isSuccess && !hasAddress && !isAutoDismissed))

  const handleClose = () => {
    setIsManualOpen(false)
    setIsAutoDismissed(true)

    if (hasAddress) {
      navigate(ROUTES.feed)
    }
  }

  return (
    <PageLayout
      className={s.root}
      header={
        hasAddress && (
          <Button
            className={s.back}
            asChild
            size="small"
            tone="ghost"
            iconBefore={<IconChevronLeft size={16} />}
          >
            <Link to={ROUTES.feed}>{commonTexts.actions.goHome}</Link>
          </Button>
        )
      }
    >
      <main className={s.hero}>
        <span className={s.illustration}>
          <OnboardingIllustration />
        </span>
        <div className={s.content}>
          <Typography.Title variant="large-strong">{onboardingTexts.title}</Typography.Title>
          <Typography.Text variant="body" color="secondary">
            {onboardingTexts.description}
          </Typography.Text>
        </div>
        <div className={s.actions}>
          <Button stretched onClick={() => setIsManualOpen(true)}>
            {onboardingTexts.action}
          </Button>
          <Typography.Text variant="note" color="tertiary">
            {onboardingTexts.consent}
          </Typography.Text>
        </div>
        <div className={s.slogan}>
          <Typography.Text variant="description" color="secondary">
            {onboardingTexts.slogan}
          </Typography.Text>
          <UkAuthEntryLink
            onClick={() => {
              setIsManualOpen(false)
              setIsAutoDismissed(true)
              setIsUkAuthOpen(true)
            }}
          />
        </div>
      </main>
      <AddressBindSheet
        open={isAddressSheetOpen}
        onClose={handleClose}
        onSuccess={() => navigate(ROUTES.feed)}
      />
      <UkAuthSheet
        open={isUkAuthOpen}
        onClose={() => setIsUkAuthOpen(false)}
        onSuccess={() => navigate(ROUTES.dispatcher, { replace: true })}
      />
    </PageLayout>
  )
}
