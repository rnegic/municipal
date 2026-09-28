import { Typography } from '@maxhub/max-ui'

import {
  CategoryIcon,
  IncidentPhotos,
  IncidentStatusBadge,
  getIncidentLegalTerm,
  incidentTexts,
  useIncidentQuery,
  type ResidentRequest,
} from '@/entities/incident'
import { useMeQuery } from '@/entities/user'
import { IncidentConfirmButton } from '@/features/incident-confirm'
import { IconShield } from '@/shared/assets/icons'
import { formatDate, formatDateTime } from '@/shared/lib/date'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { Skeleton } from '@/shared/ui/skeleton'
import s from './RequestDetailsSheet.module.scss'

export interface RequestDetailsSheetProps {
  request: ResidentRequest | null
  onClose: () => void
}

const texts = incidentTexts.details

export const RequestDetailsSheet = ({ request, onClose }: RequestDetailsSheetProps) => {
  const incidentQuery = useIncidentQuery(request?.id)
  const meQuery = useMeQuery()

  if (request === null) {
    return null
  }

  const incident = incidentQuery.data ?? null
  const legalTerm = getIncidentLegalTerm(request.category)
  const reporterName = meQuery.data?.user.fullName ?? texts.reporterFallback
  const responsibleName = meQuery.data?.house?.uk?.name ?? texts.responsibleFallback
  const supportersCount = incident?.affectedCount ?? 0

  return (
    <BottomSheet
      open
      title={request.title}
      description={texts.sentAt(formatDateTime(request.createdAt))}
      onClose={onClose}
    >
      <div className={s.content}>
        <section className={s.status}>
          <IncidentStatusBadge status={request.status} />
          <Typography.Text variant="description" color="secondary">
            {incidentTexts.statusHints[request.status]}
          </Typography.Text>
          {request.dueAt ? (
            <Typography.Text className={s.due} variant="note" color="tertiary">
              {`${texts.dueLabel}: ${texts.dueValue(formatDate(request.dueAt))}`}
            </Typography.Text>
          ) : null}
        </section>

        <section className={s.about}>
          <div className={s.category}>
            <CategoryIcon category={request.category} size={32} className={s.categoryIcon} />
            <div className={s.categoryText}>
              <Typography.Text variant="note" color="tertiary">
                {texts.categoryLabel}
              </Typography.Text>
              <Typography.Text variant="body-strong">
                {request.category
                  ? incidentTexts.categories[request.category]
                  : texts.categoryEmpty}
              </Typography.Text>
            </div>
          </div>

          <div className={s.problem}>
            <Typography.Text variant="note" color="tertiary">
              {texts.problemLabel}
            </Typography.Text>
            {incidentQuery.isPending ? (
              <div className={s.problemSkeleton}>
                <Skeleton height={14} />
                <Skeleton height={14} width="72%" />
              </div>
            ) : incidentQuery.isError ? (
              <div className={s.error}>
                <Typography.Text variant="description" color="secondary">
                  {texts.loadError}
                </Typography.Text>
                <Button size="small" tone="secondary" onClick={() => incidentQuery.refetch()}>
                  {texts.retry}
                </Button>
              </div>
            ) : (
              <Typography.Text variant="description">{incident?.description}</Typography.Text>
            )}
          </div>

          {incident ? <IncidentPhotos photos={incident.photos} /> : null}
        </section>

        <dl className={s.facts}>
          <div className={s.fact}>
            <dt>
              <Typography.Text variant="note" color="tertiary">
                {texts.reporterLabel}
              </Typography.Text>
            </dt>
            <dd className={s.factValue}>
              <Typography.Text variant="body">{reporterName}</Typography.Text>
            </dd>
          </div>

          <div className={s.fact}>
            <dt>
              <Typography.Text variant="note" color="tertiary">
                {texts.responsibleLabel}
              </Typography.Text>
            </dt>
            <dd className={s.factValue}>
              <Typography.Text variant="body">{responsibleName}</Typography.Text>
              <Typography.Text variant="note" color="tertiary">
                {incidentTexts.authorities.UK.scope}
              </Typography.Text>
            </dd>
          </div>

          {supportersCount > 0 ? (
            <div className={s.fact}>
              <dt>
                <Typography.Text variant="note" color="tertiary">
                  {texts.supportersLabel}
                </Typography.Text>
              </dt>
              <dd className={s.factValue}>
                <Typography.Text variant="body">
                  {texts.supportersValue(supportersCount)}
                </Typography.Text>
              </dd>
            </div>
          ) : null}
        </dl>

        <section className={s.legal}>
          <header className={s.legalHead}>
            <span className={s.legalIcon} aria-hidden="true">
              <IconShield size={18} />
            </span>
            <Typography.Text variant="note-strong" color="inherit">
              {texts.legalTitle}
            </Typography.Text>
          </header>
          <Typography.Title variant="medium-strong">
            {texts.legalDays(legalTerm.days, legalTerm.businessDays)}
          </Typography.Title>
          <Typography.Text variant="description" color="secondary">
            {legalTerm.scope}
          </Typography.Text>
          <div className={s.legalFooter}>
            <Typography.Text variant="note" color="tertiary">
              {legalTerm.law}
            </Typography.Text>
          </div>
        </section>

        {request.status === 'verifying' ? (
          <IncidentConfirmButton
            incidentId={request.id}
            alreadyConfirmed={request.confirmedByMe}
            stretched
          />
        ) : null}
      </div>
    </BottomSheet>
  )
}
