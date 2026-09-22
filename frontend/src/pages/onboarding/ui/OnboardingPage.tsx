import { useState } from 'react'

import { Typography } from '@maxhub/max-ui'
import { useNavigate } from 'react-router-dom'

import { getBoundHouse } from '@/entities/house'
import { AddressBindSheet } from '@/features/address-bind'
import { IconChevronLeft } from '@/shared/assets/icons'
import { OnboardingIllustration } from '@/shared/assets/illustrations'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { Button } from '@/shared/ui/button'
import { PageLayout } from '@/shared/ui/page-layout'
import { onboardingTexts } from '../config/texts'
import s from './OnboardingPage.module.scss'

export const OnboardingPage = () => {
  const navigate = useNavigate()
  const hasBoundHouse = getBoundHouse() !== null
  const [isSheetOpen, setIsSheetOpen] = useState(!hasBoundHouse)

  const handleClose = () => {
    setIsSheetOpen(false)

    if (hasBoundHouse) {
      navigate(ROUTES.feed)
    }
  }

  return (
    <PageLayout
      className={s.root}
      header={
        hasBoundHouse && (
          <Button
            className={s.back}
            size="medium"
            tone="ghost"
            iconBefore={<IconChevronLeft size={18} />}
            onClick={() => navigate(-1)}
          >
            {commonTexts.actions.back}
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
          <Button stretched onClick={() => setIsSheetOpen(true)}>
            {onboardingTexts.action}
          </Button>
          <Typography.Text variant="note" color="tertiary">
            {onboardingTexts.consent}
          </Typography.Text>
        </div>
        <Typography.Text className={s.slogan} variant="description" color="secondary">
          {onboardingTexts.slogan}
        </Typography.Text>
      </main>
      <AddressBindSheet
        open={isSheetOpen}
        onClose={handleClose}
        onSuccess={() => navigate(ROUTES.feed)}
      />
    </PageLayout>
  )
}
