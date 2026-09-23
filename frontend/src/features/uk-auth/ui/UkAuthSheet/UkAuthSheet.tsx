import { useId, useState, type FormEvent } from 'react'

import { Input, Typography } from '@maxhub/max-ui'

import { IconShield } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { useEsiaLoginMutation } from '../../api'
import { ukAuthTexts as texts } from '../../config/texts'
import { isValidInn, normalizeInn } from '../../lib/inn'
import s from './UkAuthSheet.module.scss'

export interface UkAuthSheetProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

const ORGANIZATION_INN_LENGTH = 10

const validateInn = (value: string): string | null => {
  if (value.length !== ORGANIZATION_INN_LENGTH) {
    return texts.validation.innFormat
  }

  return isValidInn(value) ? null : texts.validation.innChecksum
}

export const UkAuthSheet = ({ open, onClose, onSuccess }: UkAuthSheetProps) => {
  const innId = useId()
  const passwordId = useId()
  const [inn, setInn] = useState('')
  const [password, setPassword] = useState('')
  const [fieldError, setFieldError] = useState<string | null>(null)
  const loginMutation = useEsiaLoginMutation()

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const innError = validateInn(inn)

    if (innError) {
      setFieldError(innError)
      return
    }

    if (!password) {
      setFieldError(texts.validation.passwordRequired)
      return
    }

    setFieldError(null)
    loginMutation.mutate(
      { inn, password },
      {
        onSuccess: () => {
          setInn('')
          setPassword('')
          onSuccess()
        },
      },
    )
  }

  const errorText = fieldError ?? (loginMutation.isError ? describeApiError(loginMutation.error).description : null)

  return (
    <BottomSheet open={open} title={texts.sheet.title} description={texts.sheet.description} onClose={onClose}>
      <form className={s.form} onSubmit={handleSubmit}>
        <div className={s.gov}>
          <span className={s.govIcon}>
            <IconShield size={22} />
          </span>
          <div className={s.govText}>
            <Typography.Text variant="body-strong">{texts.sheet.gov}</Typography.Text>
            <Typography.Text variant="note" color="secondary">
              {texts.mockNotice}
            </Typography.Text>
          </div>
        </div>

        <div className={s.field}>
          <span id={innId}>
            <Typography.Text variant="description" color="secondary">
              {texts.sheet.innLabel}
            </Typography.Text>
          </span>
          <Input
            autoFocus
            size="large"
            inputMode="numeric"
            autoComplete="off"
            maxLength={ORGANIZATION_INN_LENGTH}
            aria-labelledby={innId}
            aria-invalid={fieldError === texts.validation.innFormat || fieldError === texts.validation.innChecksum}
            value={inn}
            placeholder={texts.sheet.innPlaceholder}
            onChange={(event) => {
              setInn(normalizeInn(event.target.value))
              setFieldError(null)
            }}
          />
        </div>

        <div className={s.field}>
          <span id={passwordId}>
            <Typography.Text variant="description" color="secondary">
              {texts.sheet.passwordLabel}
            </Typography.Text>
          </span>
          <Input
            size="large"
            type="password"
            autoComplete="current-password"
            aria-labelledby={passwordId}
            aria-invalid={fieldError === texts.validation.passwordRequired}
            value={password}
            placeholder={texts.sheet.passwordPlaceholder}
            onChange={(event) => {
              setPassword(event.target.value)
              setFieldError(null)
            }}
          />
        </div>

        {errorText ? (
          <Typography.Text className={s.error} variant="note" role="alert">
            {errorText}
          </Typography.Text>
        ) : null}

        <Button type="submit" stretched disabled={loginMutation.isPending}>
          {loginMutation.isPending ? texts.sheet.submitPending : texts.sheet.submit}
        </Button>

        <Typography.Text variant="note" color="tertiary">
          {texts.disclaimer}
        </Typography.Text>
      </form>
    </BottomSheet>
  )
}
