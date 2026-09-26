import { useState } from 'react'

import { IconButton, Typography } from '@maxhub/max-ui'

import type { House } from '@/entities/house'
import { IconInfo, IconPhone } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { openLink } from '@/shared/lib/max'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { ukInfoTexts as texts } from '../../config/texts'
import s from './UkInfoButton.module.scss'

export interface UkInfoButtonProps {
  house: House
  className?: string
}

const toWebsiteUrl = (website: string): string =>
  website.startsWith('http') ? website : `https://${website}`

interface PhoneBlockProps {
  label: string
  phone: string
}

const PhoneBlock = ({ label, phone }: PhoneBlockProps) => (
  <section className={s.phoneBlock}>
    <Typography.Text variant="description" color="secondary">
      {label}
    </Typography.Text>
    <Typography.Text variant="body-strong">{phone}</Typography.Text>
    <Button
      className={s.callButton}
      size="small"
      tone="secondary"
      onClick={() => {
        window.location.href = `tel:${phone.replace(/[^\d+]/g, '')}`
      }}
    >
      <IconPhone size={16} />
      {texts.call}
    </Button>
  </section>
)

interface InfoRowProps {
  label: string
  value: string
  onClick?: () => void
}

const InfoRow = ({ label, value, onClick }: InfoRowProps) => (
  <div className={cn(s.row, onClick && s.rowClickable)} onClick={onClick}>
    <Typography.Text variant="note" color="tertiary">
      {label}
    </Typography.Text>
    <Typography.Text className={onClick ? s.rowLink : undefined} variant="body">
      {value}
    </Typography.Text>
  </div>
)

export const UkInfoButton = ({ house, className }: UkInfoButtonProps) => {
  const [open, setOpen] = useState(false)
  const uk = house.uk

  if (!uk) {
    return null
  }

  const { phone, emergencyPhone, email, website, officeAddress, workingHours } = uk

  const hasContacts =
    phone !== null ||
    emergencyPhone !== null ||
    email !== null ||
    website !== null ||
    officeAddress !== null ||
    workingHours !== null

  return (
    <>
      <IconButton
        size="small"
        variant="secondary"
        className={className}
        aria-label={texts.triggerLabel}
        onClick={() => setOpen(true)}
      >
        <IconInfo />
      </IconButton>
      <BottomSheet open={open} title={texts.title} description={uk.name} onClose={() => setOpen(false)}>
        <div className={s.content}>
          {phone !== null ? <PhoneBlock label={texts.phone} phone={phone} /> : null}
          {emergencyPhone !== null ? (
            <PhoneBlock label={texts.emergencyPhone} phone={emergencyPhone} />
          ) : null}
          {hasContacts ? (
            <div className={s.rows}>
              {workingHours !== null ? (
                <InfoRow label={texts.workingHours} value={workingHours} />
              ) : null}
              {officeAddress !== null ? (
                <InfoRow label={texts.officeAddress} value={officeAddress} />
              ) : null}
              {website !== null ? (
                <InfoRow
                  label={texts.website}
                  value={website}
                  onClick={() => openLink(toWebsiteUrl(website))}
                />
              ) : null}
              {email !== null ? (
                <InfoRow
                  label={texts.email}
                  value={email}
                  onClick={() => {
                    window.location.href = `mailto:${email}`
                  }}
                />
              ) : null}
            </div>
          ) : (
            <Typography.Text variant="description" color="secondary">
              {texts.empty}
            </Typography.Text>
          )}
        </div>
      </BottomSheet>
    </>
  )
}
