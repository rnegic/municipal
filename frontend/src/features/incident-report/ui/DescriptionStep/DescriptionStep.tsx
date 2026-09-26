import { useId, type FormEvent } from 'react'

import { Input, Textarea, Typography } from '@maxhub/max-ui'

import {
  INCIDENT_DESCRIPTION_MAX_LENGTH,
  INCIDENT_DESCRIPTION_MIN_LENGTH,
  INCIDENT_TITLE_MAX_LENGTH,
  INCIDENT_TITLE_MIN_LENGTH,
} from '@/entities/incident'
import { Button } from '@/shared/ui/button'
import { LoadingState } from '@/shared/ui/loading-state'
import { Skeleton } from '@/shared/ui/skeleton'
import { incidentReportTexts as texts } from '../../config/texts'
import s from './DescriptionStep.module.scss'

export interface DescriptionStepProps {
  title: string
  value: string
  onTitleChange: (value: string) => void
  onChange: (value: string) => void
  onSubmit: () => void
  onManual: () => void
  isAnalyzing: boolean
  validationError: string | null
}

export const DescriptionStep = ({
  title,
  value,
  onTitleChange,
  onChange,
  onSubmit,
  onManual,
  isAnalyzing,
  validationError,
}: DescriptionStepProps) => {
  const titleId = useId()
  const fieldId = useId()
  const errorId = useId()

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit()
  }

  if (isAnalyzing) {
    return (
      <div className={s.analyzing} aria-live="polite">
        <div className={s.analyzingHeading}>
          <Typography.Title variant="medium-strong">
            {texts.descriptionStep.analyzingTitle}
          </Typography.Title>
          <Typography.Text variant="description" color="secondary">
            {texts.descriptionStep.analyzingDescription}
          </Typography.Text>
        </div>
        <Skeleton height={96} radius="card" />
        <Skeleton height={20} width="70%" />
        <Skeleton height={20} width="45%" />
        <LoadingState label={texts.descriptionStep.analyzingTitle} />
      </div>
    )
  }

  return (
    <form className={s.form} onSubmit={handleSubmit} noValidate>
      <div className={s.field}>
        <label htmlFor={titleId}>
          <Typography.Text variant="description" color="secondary">
            {texts.descriptionStep.titleLabel}
          </Typography.Text>
        </label>
        <Input
          id={titleId}
          className={s.titleInput}
          size="large"
          maxLength={INCIDENT_TITLE_MAX_LENGTH}
          placeholder={texts.descriptionStep.titlePlaceholder}
          value={title}
          aria-invalid={Boolean(validationError)}
          aria-describedby={validationError ? errorId : undefined}
          onChange={(event) => onTitleChange(event.target.value)}
        />
        <div className={s.meta}>
          <Typography.Text variant="note" color="tertiary">
            {texts.descriptionStep.hint(INCIDENT_TITLE_MIN_LENGTH, INCIDENT_TITLE_MAX_LENGTH)}
          </Typography.Text>
          <Typography.Text variant="note" color="tertiary">
            {texts.descriptionStep.counter(title.length, INCIDENT_TITLE_MAX_LENGTH)}
          </Typography.Text>
        </div>
      </div>

      <div className={s.field}>
        <label htmlFor={fieldId}>
          <Typography.Text variant="description" color="secondary">
            {texts.descriptionStep.label}
          </Typography.Text>
        </label>
        <Textarea
          id={fieldId}
          className={s.textareaRoot}
          innerClassNames={{ textarea: s.textarea }}
          value={value}
          rows={5}
          maxLength={INCIDENT_DESCRIPTION_MAX_LENGTH}
          placeholder={texts.descriptionStep.placeholder}
          aria-invalid={Boolean(validationError)}
          aria-describedby={validationError ? errorId : undefined}
          onChange={(event) => onChange(event.target.value)}
        />
        <div className={s.meta}>
          <Typography.Text variant="note" color="tertiary">
            {texts.descriptionStep.hint(
              INCIDENT_DESCRIPTION_MIN_LENGTH,
              INCIDENT_DESCRIPTION_MAX_LENGTH,
            )}
          </Typography.Text>
          <Typography.Text variant="note" color="tertiary">
            {texts.descriptionStep.counter(value.length, INCIDENT_DESCRIPTION_MAX_LENGTH)}
          </Typography.Text>
        </div>
      </div>

      {validationError ? (
        <span id={errorId} className={s.error} role="alert">
          <Typography.Text variant="note" color="inherit">
            {validationError}
          </Typography.Text>
        </span>
      ) : null}

      <div className={s.actions}>
        <Button type="submit" stretched>
          {texts.descriptionStep.submit}
        </Button>
        <Button type="button" tone="secondary" stretched onClick={onManual}>
          {texts.descriptionStep.manualSubmit}
        </Button>
      </div>
    </form>
  )
}
