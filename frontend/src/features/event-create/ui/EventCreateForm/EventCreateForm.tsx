import { useState, type FormEvent } from 'react'

import { Typography } from '@maxhub/max-ui'
import dayjs from 'dayjs'

import { useCreateEventMutation, type CreateEventInput } from '@/entities/event'
import type { House } from '@/entities/house'
import { describeApiError } from '@/shared/lib/api-error'
import { Button } from '@/shared/ui/button'
import { eventCreateTexts as texts } from '../../config/texts'
import s from './EventCreateForm.module.scss'

export interface EventCreateFormProps {
  houses: readonly House[]
  onSuccess: () => void
  onCancel?: () => void
}

export const EventCreateForm = ({ houses, onSuccess, onCancel }: EventCreateFormProps) => {
  const createMutation = useCreateEventMutation()
  const [houseId, setHouseId] = useState(houses[0]?.id ?? '')
  const [reason, setReason] = useState('')
  const [responsible, setResponsible] = useState('')
  const [entrance, setEntrance] = useState('')
  const [riser, setRiser] = useState('')
  const [scheduledFrom, setScheduledFrom] = useState('')
  const [scheduledTo, setScheduledTo] = useState('')
  const [validationError, setValidationError] = useState<string | null>(null)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (!houseId || !reason.trim() || !responsible.trim() || !scheduledFrom || !scheduledTo) {
      setValidationError(texts.required)
      return
    }

    const input: CreateEventInput = {
      houseId,
      reason: reason.trim(),
      responsible: responsible.trim(),
      entrance: entrance.trim() || undefined,
      riser: riser.trim() || undefined,
      scheduledFrom: dayjs(scheduledFrom).format(),
      scheduledTo: dayjs(scheduledTo).format(),
    }

    if (!dayjs(input.scheduledTo).isAfter(dayjs(input.scheduledFrom))) {
      setValidationError(texts.invalidRange)
      return
    }

    setValidationError(null)

    try {
      await createMutation.mutateAsync(input)
      onSuccess()
    } catch {
      setValidationError(null)
    }
  }

  return (
    <form className={s.form} onSubmit={handleSubmit}>
      <label className={s.field}>
        <Typography.Text variant="description" color="secondary">
          {texts.houseLabel}
        </Typography.Text>
        <select className={s.select} value={houseId} onChange={(event) => setHouseId(event.target.value)}>
          {houses.map((house) => (
            <option key={house.id} value={house.id}>
              {house.address}
            </option>
          ))}
        </select>
      </label>
      <label className={s.field}>
        <Typography.Text variant="description" color="secondary">
          {texts.reasonLabel}
        </Typography.Text>
        <input
          className={s.input}
          maxLength={120}
          placeholder={texts.reasonPlaceholder}
          value={reason}
          onChange={(event) => setReason(event.target.value)}
        />
      </label>
      <label className={s.field}>
        <Typography.Text variant="description" color="secondary">
          {texts.responsibleLabel}
        </Typography.Text>
        <input
          className={s.input}
          maxLength={80}
          placeholder={texts.responsiblePlaceholder}
          value={responsible}
          onChange={(event) => setResponsible(event.target.value)}
        />
      </label>
      <label className={s.field}>
        <Typography.Text variant="description" color="secondary">
          {texts.entranceLabel}
        </Typography.Text>
        <input
          className={s.input}
          maxLength={40}
          placeholder={texts.entrancePlaceholder}
          value={entrance}
          onChange={(event) => setEntrance(event.target.value)}
        />
      </label>
      <label className={s.field}>
        <Typography.Text variant="description" color="secondary">
          {texts.riserLabel}
        </Typography.Text>
        <input
          className={s.input}
          maxLength={120}
          placeholder={texts.riserPlaceholder}
          value={riser}
          onChange={(event) => setRiser(event.target.value)}
        />
      </label>
      <div className={s.row}>
        <label className={s.field}>
          <Typography.Text variant="description" color="secondary">
            {texts.fromLabel}
          </Typography.Text>
          <input
            className={s.input}
            type="datetime-local"
            value={scheduledFrom}
            onChange={(event) => setScheduledFrom(event.target.value)}
          />
        </label>
        <label className={s.field}>
          <Typography.Text variant="description" color="secondary">
            {texts.toLabel}
          </Typography.Text>
          <input
            className={s.input}
            type="datetime-local"
            value={scheduledTo}
            onChange={(event) => setScheduledTo(event.target.value)}
          />
        </label>
      </div>
      {validationError ? (
        <Typography.Text className={s.error} variant="note">
          {validationError}
        </Typography.Text>
      ) : null}
      {createMutation.error ? (
        <Typography.Text className={s.error} variant="note">
          {describeApiError(createMutation.error).description}
        </Typography.Text>
      ) : null}
      <div className={s.actions}>
        <Button type="submit" stretched disabled={createMutation.isPending}>
          {texts.submit}
        </Button>
        {onCancel ? (
          <Button
            type="button"
            tone="secondary"
            stretched
            onClick={onCancel}
            disabled={createMutation.isPending}
          >
            {texts.cancel}
          </Button>
        ) : null}
      </div>
    </form>
  )
}
