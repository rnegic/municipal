import { Typography } from '@maxhub/max-ui'

import { incidentTexts, type IncidentAuthority, type IncidentCategory } from '@/entities/incident'
import { IconAlertTriangle, IconCheckCircle } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { incidentReportTexts as texts } from '../../config/texts'
import s from './RoutingVerdict.module.scss'

export interface RoutingVerdictData {
  authority: IncidentAuthority
  isUkResponsibility: boolean
  category?: IncidentCategory | null
  reasoningText?: string | null
}

export interface RoutingVerdictProps {
  decision: RoutingVerdictData
  className?: string
}

export const RoutingVerdict = ({ decision, className }: RoutingVerdictProps) => {
  const isUk = decision.isUkResponsibility
  const authority = incidentTexts.authorities[decision.authority]

  return (
    <section
      className={cn(s.root, isUk ? s.toneUk : s.toneOther, className)}
      aria-live="polite"
    >
      <header className={s.header}>
        <span className={s.icon} aria-hidden="true">
          {isUk ? <IconCheckCircle size={22} /> : <IconAlertTriangle size={22} />}
        </span>
        <Typography.Title variant="medium-strong">
          {isUk ? texts.routingStep.ukTitle : texts.routingStep.otherTitle}
        </Typography.Title>
      </header>

      <dl className={s.facts}>
        {decision.category ? (
          <div className={s.fact}>
            <dt>
              <Typography.Text variant="note" color="inherit">
                {texts.routingStep.categoryLabel}
              </Typography.Text>
            </dt>
            <dd className={s.factValue}>
              <Typography.Text variant="body-strong" color="inherit">
                {incidentTexts.categories[decision.category]}
              </Typography.Text>
            </dd>
          </div>
        ) : null}
        <div className={s.fact}>
          <dt>
            <Typography.Text variant="note" color="inherit">
              {texts.routingStep.authorityLabel}
            </Typography.Text>
          </dt>
          <dd className={s.factValue}>
            <Typography.Text variant="body-strong" color="inherit">
              {authority.name}
            </Typography.Text>
            <Typography.Text variant="note" color="inherit">
              {authority.scope}
            </Typography.Text>
          </dd>
        </div>
      </dl>

      <Typography.Text variant="description" color="inherit">
        {isUk ? texts.routingStep.ukDescription : texts.routingStep.otherDescription}
      </Typography.Text>

      {decision.reasoningText ? (
        <Typography.Text className={s.reasoning} variant="note" color="inherit">
          {decision.reasoningText}
        </Typography.Text>
      ) : null}
    </section>
  )
}
