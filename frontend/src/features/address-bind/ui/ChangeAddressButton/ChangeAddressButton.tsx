import { useState } from 'react'

import { IconButton } from '@maxhub/max-ui'

import { IconPencil } from '@/shared/assets/icons'
import { commonTexts } from '@/shared/config/texts'
import { AddressBindSheet } from '../AddressBindSheet'

export const ChangeAddressButton = () => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <IconButton
        size="small"
        variant="secondary"
        aria-label={commonTexts.actions.changeAddress}
        onClick={() => setOpen(true)}
      >
        <IconPencil />
      </IconButton>
      <AddressBindSheet open={open} onClose={() => setOpen(false)} onSuccess={() => setOpen(false)} />
    </>
  )
}
