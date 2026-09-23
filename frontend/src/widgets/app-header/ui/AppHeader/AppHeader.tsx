import { Typography } from '@maxhub/max-ui'

import { formatHouseAddress, type House } from '@/entities/house'
import { ChangeAddressButton } from '@/features/address-bind'
import { SignOutButton } from '@/features/sign-out'
import { ThemeToggle } from '@/features/theme-switch'
import { IconLocation } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { Card } from '@/shared/ui/card'
import { StatusDot, type StatusTone } from '@/shared/ui/status-dot'
import s from './AppHeader.module.scss'

export interface AppHeaderStatus {
  tone: StatusTone
  label: string
}

export interface AppHeaderProps {
  house: House
  status: AppHeaderStatus
  className?: string
}

export const AppHeader = ({ house, status, className }: AppHeaderProps) => (
  <Card className={cn(s.root, className)} padding="compact">
    <div className={s.addressRow}>
      <span className={s.addressIcon}>
        <IconLocation size={24} />
      </span>
      <div className={s.address}>
        <Typography.Title variant="small-strong">{formatHouseAddress(house)}</Typography.Title>
      </div>
      <div className={s.actions}>
        <ThemeToggle />
        <ChangeAddressButton />
        <SignOutButton />
      </div>
    </div>
    <div className={s.statusRow}>
      <StatusDot tone={status.tone} size="medium" pulse={status.tone === 'danger'} />
      <Typography.Text variant="description" color="secondary">
        {status.label}
      </Typography.Text>
    </div>
  </Card>
)
