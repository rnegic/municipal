import { useState } from 'react'

import { Input, Typography } from '@maxhub/max-ui'

import { saveBoundHouse } from '@/entities/house'
import { IconLocation } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { useBindHouseMutation } from '../../api'
import { addressBindTexts } from '../../config/texts'
import s from './AddressBindSheet.module.scss'

export interface AddressBindSheetProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

const MIN_ADDRESS_LENGTH = 5

export const AddressBindSheet = ({ open, onClose, onSuccess }: AddressBindSheetProps) => {
  const [address, setAddress] = useState('')
  const [invalid, setInvalid] = useState(false)
  const bindMutation = useBindHouseMutation()

  const handleSubmit = (event: React.SubmitEvent) => {
    event.preventDefault()

    const value = address.trim()

    if (value.length < MIN_ADDRESS_LENGTH) {
      setInvalid(true)
      return
    }

    bindMutation.mutate(value, {
      onSuccess: (house) => {
        saveBoundHouse(house)
        onSuccess()
      },
    })
  }

  return (
    <BottomSheet
      open={open}
      title={addressBindTexts.title}
      description={addressBindTexts.description}
      onClose={onClose}
    >
      <form className={s.form} onSubmit={handleSubmit}>
        <label className={s.field}>
          <Typography.Text variant="description" color="secondary">
            {addressBindTexts.label}
          </Typography.Text>
          <Input
            autoFocus
            withClearButton
            size="large"
            autoComplete="street-address"
            value={address}
            placeholder={addressBindTexts.placeholder}
            iconBefore={<IconLocation />}
            aria-invalid={invalid}
            hint={invalid ? addressBindTexts.validation : undefined}
            onChange={(event) => {
              setAddress(event.target.value)
              setInvalid(false)
            }}
          />
        </label>
        {bindMutation.isError ? (
          <Typography.Text variant="note" color="secondary">
            {describeApiError(bindMutation.error).description}
          </Typography.Text>
        ) : null}
        <Button type="submit" stretched disabled={bindMutation.isPending}>
          {addressBindTexts.submit}
        </Button>
      </form>
    </BottomSheet>
  )
}
