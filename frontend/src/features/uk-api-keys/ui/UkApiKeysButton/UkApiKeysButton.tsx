import { IconButton } from '@maxhub/max-ui'
import { useState } from 'react'

import { IconKey } from '@/shared/assets/icons'
import { ukApiKeysTexts as texts } from '../../config/texts'
import { UkApiKeysSheet } from '../UkApiKeysSheet'

export const UkApiKeysButton = () => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <IconButton size="small" variant="secondary" aria-label={texts.open} onClick={() => setOpen(true)}>
        <IconKey />
      </IconButton>
      <UkApiKeysSheet open={open} onClose={() => setOpen(false)} />
    </>
  )
}
