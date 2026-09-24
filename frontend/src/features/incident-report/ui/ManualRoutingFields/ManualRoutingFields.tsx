import { Typography } from '@maxhub/max-ui'

import {
  incidentAuthorityOptions,
  incidentCategoryOptions,
  type IncidentAuthority,
  type IncidentCategory,
} from '@/entities/incident'
import { Select } from '@/shared/ui/select'
import { incidentReportTexts as texts } from '../../config/texts'
import s from './ManualRoutingFields.module.scss'

export interface ManualRoutingFieldsProps {
  authority: IncidentAuthority | ''
  category: IncidentCategory | ''
  onAuthorityChange: (value: IncidentAuthority) => void
  onCategoryChange: (value: IncidentCategory) => void
  disabled?: boolean
}

export const ManualRoutingFields = ({
  authority,
  category,
  onAuthorityChange,
  onCategoryChange,
  disabled = false,
}: ManualRoutingFieldsProps) => (
  <section className={s.root} aria-live="polite">
    <div className={s.heading}>
      <Typography.Title variant="medium-strong">{texts.fallback.title}</Typography.Title>
      <Typography.Text variant="description" color="secondary">
        {texts.fallback.description}
      </Typography.Text>
    </div>

    <Select
      label={texts.fallback.authorityLabel}
      placeholder={texts.fallback.authorityPlaceholder}
      options={incidentAuthorityOptions}
      value={authority}
      disabled={disabled}
      onValueChange={onAuthorityChange}
    />

    <Select
      label={texts.fallback.categoryLabel}
      placeholder={texts.fallback.categoryPlaceholder}
      options={incidentCategoryOptions}
      value={category}
      disabled={disabled}
      onValueChange={onCategoryChange}
    />
  </section>
)
