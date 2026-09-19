import { Link } from 'react-router-dom'

import { IconButton, Typography } from '@maxhub/max-ui'

import { formatHouseAddress, type House } from '@/entities/house'
import { ThemeToggle } from '@/features/theme-switch'
import { IconLocation, IconPencil } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
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
        <IconButton size="small" variant="secondary" asChild>
          <Link to={ROUTES.onboarding} aria-label={commonTexts.actions.changeAddress}>
            <IconPencil />
          </Link>
        </IconButton>
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
