import { useId } from 'react'

import { Typography } from '@maxhub/max-ui'

import {
  INCIDENT_PHOTOS_MAX_COUNT,
  INCIDENT_PHOTO_MAX_BYTES,
  INCIDENT_PHOTO_MIME_TYPES,
} from '@/entities/incident'
import { describeApiError } from '@/shared/lib/api-error'
import { Button } from '@/shared/ui/button'
import { PhotoPicker, type PhotoPickerItem } from '@/shared/ui/photo-picker'
import { incidentReportTexts as texts } from '../../config/texts'
import s from './PhotoStep.module.scss'

export interface PhotoStepProps {
  photos: PhotoPickerItem[]
  isPhotoRequired: boolean
  isSubmitting: boolean
  validationError: string | null
  submitError: unknown
  onChange: (photos: PhotoPickerItem[]) => void
  onSubmit: () => void
  onBack: () => void
}

export const PhotoStep = ({
  photos,
  isPhotoRequired,
  isSubmitting,
  validationError,
  submitError,
  onChange,
  onSubmit,
  onBack,
}: PhotoStepProps) => {
  const errorId = useId()

  const isSubmitDisabled = isSubmitting || (isPhotoRequired && photos.length === 0)

  const handleSubmit = (event: React.SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit()
  }

  return (
    <form className={s.form} onSubmit={handleSubmit} noValidate>
      <div className={s.heading}>
        <Typography.Title variant="medium-strong">
          {isPhotoRequired ? texts.photoStep.requiredTitle : texts.photoStep.optionalTitle}
        </Typography.Title>
        <Typography.Text variant="description" color="secondary">
          {isPhotoRequired
            ? texts.photoStep.requiredDescription
            : texts.photoStep.optionalDescription}
        </Typography.Text>
      </div>

      <PhotoPicker
        items={photos}
        maxFiles={INCIDENT_PHOTOS_MAX_COUNT}
        maxBytes={INCIDENT_PHOTO_MAX_BYTES}
        acceptedMimeTypes={INCIDENT_PHOTO_MIME_TYPES}
        disabled={isSubmitting}
        onChange={onChange}
      />

      {validationError ? (
        <span id={errorId} className={s.error} role="alert">
          <Typography.Text variant="note" color="inherit">
            {validationError}
          </Typography.Text>
        </span>
      ) : null}

      {submitError ? (
        <span className={s.error} role="alert">
          <Typography.Text variant="note" color="inherit">
            {describeApiError(submitError).description}
          </Typography.Text>
        </span>
      ) : null}

      <div className={s.actions}>
        <Button type="submit" stretched disabled={isSubmitDisabled}>
          {texts.photoStep.submit}
        </Button>
        <Button type="button" tone="secondary" stretched disabled={isSubmitting} onClick={onBack}>
          {texts.photoStep.back}
        </Button>
      </div>
    </form>
  )
}
