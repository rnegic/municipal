import { useCurrentHouseQuery } from '@/entities/user'
import { IconQrCode } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { openLink } from '@/shared/lib/max'
import { Button } from '@/shared/ui/button'
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
    <div className={cn(s.root, className)}>
      <span className={s.icon} aria-hidden="true">
        <IconQrCode size={20} />
      </span>
      <Button size="small" tone="secondary" onClick={() => openLink(buildHouseStickerUrl(house))}>
        {texts.action}
      </Button>
    </div>
  )
}
