import { useId } from 'react'

import { Input, Typography } from '@maxhub/max-ui'

import {
  INCIDENT_ENTRANCE_MAX_LENGTH,
  INCIDENT_FLOOR_ZONE_MAX_LENGTH,
  isUkAuthority,
  type IncidentAuthority,
  type IncidentCategory,
} from '@/entities/incident'
import { Button } from '@/shared/ui/button'
import { GOSUSLUGI_POS_URL } from '../../config/constants'
import { incidentReportTexts as texts } from '../../config/texts'
import type { IncidentRoutingDecision } from '../../model/types'
import { ManualRoutingFields } from '../ManualRoutingFields'
import { RoutingVerdict } from '../RoutingVerdict'
import s from './RoutingStep.module.scss'

export interface RoutingStepProps {
  decision: IncidentRoutingDecision | null
  isManualRouting: boolean
  manualAuthority: IncidentAuthority | ''
  manualCategory: IncidentCategory | ''
  entrance: string
  floorZone: string
  validationError: string | null
  onManualAuthorityChange: (value: IncidentAuthority) => void
  onManualCategoryChange: (value: IncidentCategory) => void
  onEntranceChange: (value: string) => void
  onFloorZoneChange: (value: string) => void
  onSubmit: () => void
  onBack: () => void
  onManualOverride: () => void
}

export const RoutingStep = ({
  decision,
  isManualRouting,
  manualAuthority,
  manualCategory,
  entrance,
  floorZone,
  validationError,
  onManualAuthorityChange,
  onManualCategoryChange,
  onEntranceChange,
  onFloorZoneChange,
  onSubmit,
  onBack,
  onManualOverride,
}: RoutingStepProps) => {
  const entranceId = useId()
  const floorZoneId = useId()
  const errorId = useId()

  const isUkFlow = decision?.isUkResponsibility ?? false

  // В ручном сценарии показываем вердикт сразу после выбора ведомства,
  // не дожидаясь выбора категории.
  const verdict =
    decision ??
    (isManualRouting && manualAuthority
      ? { authority: manualAuthority, isUkResponsibility: isUkAuthority(manualAuthority) }
      : null)

  const handleSubmit = (event: React.SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    onSubmit()
  }

  return (
    <form className={s.form} onSubmit={handleSubmit} noValidate>
      {isManualRouting ? (
        <ManualRoutingFields
          authority={manualAuthority}
          category={manualCategory}
          onAuthorityChange={onManualAuthorityChange}
          onCategoryChange={onManualCategoryChange}
        />
      ) : null}

      {verdict ? <RoutingVerdict decision={verdict} /> : null}

      {decision && !isManualRouting ? (
        <div className={s.override}>
          <Typography.Text variant="note" color="tertiary">
            {texts.routingStep.mayBeWrongHint}
          </Typography.Text>
          <Button
            className={s.overrideButton}
            type="button"
            size="small"
            tone="secondary"
            stretched
            onClick={onManualOverride}
          >
            {texts.routingStep.overrideAction}
          </Button>
        </div>
      ) : null}

      {isUkFlow ? (
        <fieldset className={s.details}>
          <legend className={s.legend}>
            <Typography.Text variant="note-strong">
              {texts.routingStep.detailsTitle}
            </Typography.Text>
          </legend>
          <div className={s.field}>
            <label htmlFor={entranceId}>
              <Typography.Text variant="description" color="secondary">
                {texts.routingStep.entranceLabel}
              </Typography.Text>
            </label>
            <Input
              id={entranceId}
              size="large"
              inputMode="numeric"
              autoComplete="off"
              maxLength={INCIDENT_ENTRANCE_MAX_LENGTH}
              placeholder={texts.routingStep.entrancePlaceholder}
              value={entrance}
              onChange={(event) => onEntranceChange(event.target.value)}
            />
          </div>
          <div className={s.field}>
            <label htmlFor={floorZoneId}>
              <Typography.Text variant="description" color="secondary">
                {texts.routingStep.floorZoneLabel}
              </Typography.Text>
            </label>
            <Input
              id={floorZoneId}
              size="large"
              autoComplete="off"
              maxLength={INCIDENT_FLOOR_ZONE_MAX_LENGTH}
              placeholder={texts.routingStep.floorZonePlaceholder}
              value={floorZone}
              onChange={(event) => onFloorZoneChange(event.target.value)}
            />
          </div>
        </fieldset>
      ) : null}

      {validationError ? (
        <span id={errorId} className={s.error} role="alert">
          <Typography.Text variant="note" color="inherit">
            {validationError}
          </Typography.Text>
        </span>
      ) : null}

      <div className={s.actions}>
        {isUkFlow ? (
          <Button type="submit" stretched>
            {texts.routingStep.submit}
          </Button>
        ) : null}

        {decision && !isUkFlow ? (
          <Button asChild stretched>
            <a href={GOSUSLUGI_POS_URL} target="_blank" rel="noopener noreferrer">
              {texts.routingStep.gosuslugiAction}
            </a>
          </Button>
        ) : null}

        <Button type="button" tone="secondary" stretched onClick={onBack}>
          {texts.routingStep.back}
        </Button>
      </div>
    </form>
  )
}
