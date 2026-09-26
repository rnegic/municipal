import { Typography } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import { incidentReportTexts as texts } from '../../config/texts'
import { useIncidentReportWizard } from '../../model/use-incident-report-wizard'
import { DescriptionStep } from '../DescriptionStep'
import { PhotoStep } from '../PhotoStep'
import { RoutingStep } from '../RoutingStep'
import s from './IncidentReportForm.module.scss'

export interface IncidentReportFormProps {
  onSuccess: () => void
  onCancel?: () => void
  className?: string
}

export const IncidentReportForm = ({ onSuccess, onCancel, className }: IncidentReportFormProps) => {
  const wizard = useIncidentReportWizard({ onSuccess })

  return (
    <div className={cn(s.root, className)}>
      <Typography.Text className={s.stepLabel} variant="note-strong">
        {texts.steps[wizard.step]}
      </Typography.Text>

      {wizard.step === 'description' ? (
        <DescriptionStep
          title={wizard.title}
          value={wizard.description}
          isAnalyzing={wizard.isAnalyzing}
          validationError={wizard.validationError}
          onTitleChange={wizard.setTitle}
          onChange={wizard.setDescription}
          onSubmit={wizard.analyze}
          onManual={wizard.startManualRouting}
        />
      ) : null}

      {wizard.step === 'routing' ? (
        <RoutingStep
          decision={wizard.decision}
          isManualRouting={wizard.isManualRouting}
          manualAuthority={wizard.manualAuthority}
          manualCategory={wizard.manualCategory}
          entrance={wizard.entrance}
          floorZone={wizard.floorZone}
          validationError={wizard.validationError}
          onManualAuthorityChange={wizard.setManualAuthority}
          onManualCategoryChange={wizard.setManualCategory}
          onEntranceChange={wizard.setEntrance}
          onFloorZoneChange={wizard.setFloorZone}
          onSubmit={wizard.goToPhoto}
          onBack={wizard.goToDescription}
          onManualOverride={wizard.switchToManualRouting}
        />
      ) : null}

      {wizard.step === 'photo' && wizard.decision ? (
        <PhotoStep
          photos={wizard.photos}
          isPhotoRequired={wizard.decision.photoRequired}
          isSubmitting={wizard.isSubmitting}
          validationError={wizard.validationError}
          submitError={wizard.submitError}
          onChange={wizard.setPhotos}
          onSubmit={wizard.submit}
          onBack={wizard.goToRouting}
        />
      ) : null}

      {onCancel ? (
        <Button
          type="button"
          tone="secondary"
          stretched
          disabled={wizard.isAnalyzing || wizard.isSubmitting}
          onClick={onCancel}
        >
          {texts.cancel}
        </Button>
      ) : null}
    </div>
  )
}
