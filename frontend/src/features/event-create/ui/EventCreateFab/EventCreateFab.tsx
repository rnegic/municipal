import { useState } from 'react'

import type { House } from '@/entities/house'
import { IconPlus } from '@/shared/assets/icons'
import { Button } from '@/shared/ui/button'
import { eventCreateTexts } from '../../config/texts'
import { EventCreateSheet } from '../EventCreateSheet'
import s from './EventCreateFab.module.scss'

export interface EventCreateFabProps {
  houses: readonly House[]
}

export const EventCreateFab = ({ houses }: EventCreateFabProps) => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        className={s.root}
        size="large"
        stretched
        tone="secondary"
        iconBefore={<IconPlus />}
        onClick={() => setOpen(true)}
      >
        {eventCreateTexts.action}
      </Button>
      <EventCreateSheet open={open} onClose={() => setOpen(false)} houses={houses} />
    </>
  )
}
