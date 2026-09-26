import { Typography } from '@maxhub/max-ui'

import { useCurrentHouseQuery } from '@/entities/user'
import { IconQrCode } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { openLink } from '@/shared/lib/max'
import { Button } from '@/shared/ui/button'
import { Card } from '@/shared/ui/card'
import { houseStickerTexts as texts } from '../../config/texts'
import { buildHouseStickerUrl } from '../../lib/sticker-url'
import s from './HouseStickerCard.module.scss'

export interface HouseStickerCardProps {
  className?: string
}

export const HouseStickerCard = ({ className }: HouseStickerCardProps) => {
  const houseQuery = useCurrentHouseQuery()
  const house = houseQuery.data

  if (!house) {
    return null
  }

  return (
    <Card className={cn(s.root, className)} padding="compact">
      <div className={s.head}>
        <span className={s.icon} aria-hidden="true">
          <IconQrCode size={24} />
        </span>
        <div className={s.text}>
          <Typography.Text variant="body-strong">{texts.title}</Typography.Text>
          <Typography.Text variant="note" color="secondary">
            {texts.description}
          </Typography.Text>
        </div>
      </div>
      <Button
        className={s.action}
        size="small"
        tone="secondary"
        stretched
        onClick={() => openLink(buildHouseStickerUrl(house))}
      >
        {texts.action}
      </Button>
      <Typography.Text className={s.hint} variant="note" color="tertiary">
        {texts.hint}
      </Typography.Text>
    </Card>
  )
}
