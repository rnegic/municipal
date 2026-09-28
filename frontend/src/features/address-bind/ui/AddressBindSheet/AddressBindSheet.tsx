import { useId, useState } from 'react'

import { IconButton, Input, Typography } from '@maxhub/max-ui'

import { saveBoundHouse } from '@/entities/house'
import { IconLocation, IconMyLocation } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { fetchAddressByCoords } from '@/shared/lib/dadata'
import {
  GeolocationError,
  getCurrentPosition,
  type GeolocationErrorCode,
} from '@/shared/lib/geolocation'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { useAddressSuggestionsQuery, useBindHouseMutation } from '../../api'
import { addressBindTexts } from '../../config/texts'
import { AddressSuggestions } from '../AddressSuggestions'
import s from './AddressBindSheet.module.scss'

export interface AddressBindSheetProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

const MIN_ADDRESS_LENGTH = 5
const GEOCODE_COUNT = 1

export const AddressBindSheet = ({ open, onClose, onSuccess }: AddressBindSheetProps) => {
  const captionId = useId()
  const listId = useId()
  const [address, setAddress] = useState('')
  const [invalid, setInvalid] = useState(false)
  const [isSuggestOpen, setIsSuggestOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)
  const [isLocating, setIsLocating] = useState(false)
  const [locationError, setLocationError] = useState<GeolocationErrorCode | 'empty' | null>(null)
  const bindMutation = useBindHouseMutation()
  const { suggestions, isFetching, isSettled } = useAddressSuggestionsQuery(isSuggestOpen ? address : '')

  const showSuggestions = isSuggestOpen && suggestions.length > 0

  const applySuggestion = (value: string) => {
    setAddress(value)
    setInvalid(false)
    setIsSuggestOpen(false)
    setActiveIndex(-1)
    setLocationError(null)
  }

  const handleChange = (value: string) => {
    setAddress(value)
    setInvalid(false)
    setLocationError(null)
    setIsSuggestOpen(true)
    setActiveIndex(-1)
  }

  const handleLocate = async () => {
    if (isLocating) {
      return
    }

    setIsLocating(true)
    setLocationError(null)

    try {
      const { lat, lon } = await getCurrentPosition()
      const { suggestions: nearest } = await fetchAddressByCoords({ lat, lon, count: GEOCODE_COUNT })

      if (nearest.length === 0) {
        setLocationError('empty')
        return
      }

      applySuggestion(nearest[0].value)
    } catch (error) {
      setLocationError(error instanceof GeolocationError ? error.code : 'unavailable')
    } finally {
      setIsLocating(false)
    }
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Escape') {
      setIsSuggestOpen(false)
      setActiveIndex(-1)
      return
    }

    if (!showSuggestions) {
      return
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveIndex((index) => (index + 1) % suggestions.length)
      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveIndex((index) => (index - 1 + suggestions.length) % suggestions.length)
      return
    }

    if (event.key === 'Enter' && activeIndex >= 0) {
      event.preventDefault()
      applySuggestion(suggestions[activeIndex].value)
    }
  }

  const handleSubmit = (event: React.SubmitEvent) => {
    event.preventDefault()

    const value = address.trim()

    if (value.length < MIN_ADDRESS_LENGTH) {
      setInvalid(true)
      return
    }

    setIsSuggestOpen(false)

    bindMutation.mutate(value, {
      onSuccess: (house) => {
        saveBoundHouse(house)
        onSuccess()
      },
    })
  }

  const resolveStatusText = () => {
    if (showSuggestions) {
      return null
    }

    if (isFetching) {
      return addressBindTexts.suggestionsLoading
    }

    if (isSuggestOpen && isSettled) {
      return addressBindTexts.suggestionsEmpty
    }

    return null
  }

  const statusText = resolveStatusText()

  return (
    <BottomSheet
      open={open}
      title={addressBindTexts.title}
      description={addressBindTexts.description}
      onClose={onClose}
    >
      <form className={s.form} onSubmit={handleSubmit}>
        <div className={s.field}>
          <span id={captionId}>
            <Typography.Text variant="description" color="secondary">
              {addressBindTexts.label}
            </Typography.Text>
          </span>
          <Input
            autoFocus
            withClearButton
            size="large"
            innerClassNames={{ input: s.input }}
            autoComplete="street-address"
            role="combobox"
            aria-labelledby={captionId}
            aria-autocomplete="list"
            aria-expanded={showSuggestions}
            aria-controls={showSuggestions ? listId : undefined}
            aria-invalid={invalid}
            value={address}
            placeholder={addressBindTexts.placeholder}
            iconBefore={<IconLocation />}
            iconAfter={
              <IconButton
                type="button"
                variant="ghost"
                size="small"
                aria-label={addressBindTexts.locate}
                loading={isLocating}
                disabled={isLocating}
                onClick={handleLocate}
              >
                <IconMyLocation size={20} />
              </IconButton>
            }
            hint={invalid ? addressBindTexts.validation : undefined}
            onChange={(event) => handleChange(event.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={() => setIsSuggestOpen(true)}
            onBlur={() => setIsSuggestOpen(false)}
          />
          {showSuggestions ? (
            <AddressSuggestions
              listId={listId}
              suggestions={suggestions}
              activeIndex={activeIndex}
              onSelect={applySuggestion}
              onHighlight={setActiveIndex}
            />
          ) : null}
          {locationError ? (
            <Typography.Text variant="note" color="secondary">
              {locationError === 'empty'
                ? addressBindTexts.locationEmpty
                : addressBindTexts.locationErrors[locationError]}
            </Typography.Text>
          ) : null}
          {statusText ? (
            <Typography.Text variant="note" color="tertiary">
              {statusText}
            </Typography.Text>
          ) : null}
        </div>
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