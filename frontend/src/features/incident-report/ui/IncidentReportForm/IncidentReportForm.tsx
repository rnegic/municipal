import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'

import { Typography } from '@maxhub/max-ui'

import {
  useCreateIncidentMutation,
  useUploadIncidentPhotoMutation,
  type Incident,
  type IncidentSeverity,
} from '@/entities/incident'
import { describeApiError } from '@/shared/lib/api-error'
import { Button } from '@/shared/ui/button'
import { incidentReportTexts as texts } from '../../config/texts'
import s from './IncidentReportForm.module.scss'

const MAX_PHOTO_BYTES = 10 * 1024 * 1024
const ACCEPTED_PHOTO_TYPES = ['image/jpeg', 'image/png']

export interface IncidentReportFormProps {
  onSuccess: () => void
  onCancel?: () => void
}

export const IncidentReportForm = ({ onSuccess, onCancel }: IncidentReportFormProps) => {
  const createMutation = useCreateIncidentMutation()
  const uploadMutation = useUploadIncidentPhotoMutation()
  const [step, setStep] = useState<1 | 2>(1)
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [entrance, setEntrance] = useState('')
  const [riser, setRiser] = useState('')
  const [severity, setSeverity] = useState<IncidentSeverity>('critical')
  const [photo, setPhoto] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [createdIncident, setCreatedIncident] = useState<Incident | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)

  useEffect(
    () => () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl)
      }
    },
    [previewUrl],
  )

  const handlePhotoChange = (event: ChangeEvent<HTMLInputElement>) => {
    const selectedFile = event.target.files?.[0]
    event.target.value = ''

    if (!selectedFile) {
      return
    }

    if (!ACCEPTED_PHOTO_TYPES.includes(selectedFile.type)) {
      setValidationError(texts.photoType)
      return
    }

    if (selectedFile.size > MAX_PHOTO_BYTES) {
      setValidationError(texts.photoSize)
      return
    }

    setPhoto(selectedFile)
    setPreviewUrl(URL.createObjectURL(selectedFile))
    setValidationError(null)
  }

  const handleDetailsSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (!title.trim() || !description.trim()) {
      setValidationError(texts.detailsRequired)
      return
    }

    setValidationError(null)
    setStep(2)
  }

  const uploadPhoto = async (incident: Incident) => {
    if (!photo) {
      setValidationError(texts.photoRequired)
      return
    }

    try {
      await uploadMutation.mutateAsync({ incidentId: incident.id, file: photo })
      onSuccess()
    } catch {
      setValidationError(null)
    }
  }

  const handlePhotoSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (!photo) {
      setValidationError(texts.photoRequired)
      return
    }

    if (createdIncident) {
      await uploadPhoto(createdIncident)
      return
    }

    try {
      const incident = await createMutation.mutateAsync({
        title: title.trim(),
        description: description.trim(),
        severity,
        entrance: entrance.trim() || undefined,
        riser: riser.trim() || undefined,
      })
      setCreatedIncident(incident)
      await uploadPhoto(incident)
    } catch {
      setValidationError(null)
    }
  }

  const isSubmitting = createMutation.isPending || uploadMutation.isPending
  const mutationError = createMutation.error ?? uploadMutation.error

  return (
    <div className={s.root}>
      {step === 1 ? (
        <form className={s.form} onSubmit={handleDetailsSubmit}>
          <Typography.Text className={s.stepLabel} variant="note-strong">
            {texts.detailsStep}
          </Typography.Text>
          <label className={s.field}>
            <Typography.Text variant="description" color="secondary">
              {texts.titleLabel}
            </Typography.Text>
            <input
              className={s.input}
              maxLength={120}
              placeholder={texts.titlePlaceholder}
              value={title}
              onChange={(event) => setTitle(event.target.value)}
            />
          </label>
          <label className={s.field}>
            <Typography.Text variant="description" color="secondary">
              {texts.entranceLabel}
            </Typography.Text>
            <input
              className={s.input}
              maxLength={40}
              inputMode="numeric"
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
          <label className={s.field}>
            <Typography.Text variant="description" color="secondary">
              {texts.descriptionLabel}
            </Typography.Text>
            <textarea
              className={s.textarea}
              maxLength={2000}
              placeholder={texts.descriptionPlaceholder}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
          </label>
          <label className={s.field}>
            <Typography.Text variant="description" color="secondary">
              {texts.severityLabel}
            </Typography.Text>
            <select
              className={s.select}
              value={severity}
              onChange={(event) => setSeverity(event.target.value as IncidentSeverity)}
            >
              <option value="critical">{texts.critical}</option>
              <option value="warning">{texts.warning}</option>
            </select>
          </label>
          {validationError ? (
            <Typography.Text className={s.error} variant="note">
              {validationError}
            </Typography.Text>
          ) : null}
          <Button type="submit" stretched>
            {texts.next}
          </Button>
        </form>
      ) : (
        <form className={s.form} onSubmit={handlePhotoSubmit}>
          <div className={s.step}>
            <Typography.Text className={s.stepLabel} variant="note-strong">
              {texts.photoStep}
            </Typography.Text>
            <Typography.Title variant="medium-strong">{texts.photoTitle}</Typography.Title>
            <Typography.Text variant="description" color="secondary">
              {texts.photoDescription}
            </Typography.Text>
          </div>
          <div className={s.photoPicker}>
            {previewUrl ? <img className={s.preview} src={previewUrl} alt={texts.a11y.photoPreview} /> : null}
            <label>
              <input
                className={s.fileInput}
                type="file"
                accept="image/jpeg,image/png"
                capture="environment"
                onChange={handlePhotoChange}
              />
              <Button asChild tone="secondary" type="button">
                <span>{photo ? texts.changePhoto : texts.choosePhoto}</span>
              </Button>
            </label>
            {photo ? (
              <div className={s.photoMeta}>
                <Typography.Text variant="body-strong">{photo.name}</Typography.Text>
                <Typography.Text variant="note" color="tertiary">
                  {texts.photoHint}
                </Typography.Text>
              </div>
            ) : (
              <Typography.Text variant="note" color="tertiary">
                {texts.photoHint}
              </Typography.Text>
            )}
          </div>
          {validationError ? (
            <Typography.Text className={s.error} variant="note">
              {validationError}
            </Typography.Text>
          ) : null}
          {mutationError ? (
            <Typography.Text className={s.error} variant="note">
              {describeApiError(mutationError).description}
            </Typography.Text>
          ) : null}
          <div className={s.actions}>
            <Button type="submit" stretched disabled={isSubmitting}>
              {createdIncident ? texts.retryUpload : texts.upload}
            </Button>
            {!createdIncident ? (
              <Button type="button" tone="ghost" stretched onClick={() => setStep(1)} disabled={isSubmitting}>
                {texts.back}
              </Button>
            ) : null}
            {onCancel ? (
              <Button type="button" tone="ghost" stretched onClick={onCancel} disabled={isSubmitting}>
                {texts.cancel}
              </Button>
            ) : null}
          </div>
        </form>
      )}
    </div>
  )
}
