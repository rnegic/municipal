import { Typography } from '@maxhub/max-ui'
import { useMemo } from 'react'

import {
  INCIDENT_SUPPORTERS_PREVIEW_COUNT,
  incidentTexts,
  useJoinIncidentMutation,
  type IncidentSupporter,
} from '@/entities/incident'
import { IconCheckCircle, IconPlus } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { cn } from '@/shared/lib/cn'
import { AvatarStack } from '@/shared/ui/avatar-stack'
import { Button } from '@/shared/ui/button'
import { buildSupporterAvatars } from '../../lib/supporter-avatars'
import s from './IncidentJoinButton.module.scss'

export interface IncidentJoinButtonProps {
  incidentId: string
  affectedCount: number
  supporters?: readonly IncidentSupporter[]
  alreadyJoined?: boolean
  className?: string
}

export const IncidentJoinButton = ({
  incidentId,
  affectedCount,
  supporters = [],
  alreadyJoined = false,
  className,
}: IncidentJoinButtonProps) => {
  const joinMutation = useJoinIncidentMutation()
  const joined = alreadyJoined || joinMutation.isSuccess

  const avatars = useMemo(
    () => buildSupporterAvatars({ incidentId, affectedCount, supporters, joined }),
    [affectedCount, incidentId, joined, supporters],
  )

  return (
    <div className={cn(s.root, className)}>
      {affectedCount > 0 ? (
        <div className={s.confirmations}>
          <AvatarStack
            items={avatars}
            total={affectedCount}
            max={INCIDENT_SUPPORTERS_PREVIEW_COUNT}
            label={incidentTexts.a11y.supporters}
          />
          <Typography.Text variant="note" color="secondary">
            {incidentTexts.focus.affectedCount(affectedCount)}
          </Typography.Text>
        </div>
      ) : null}
      {joinMutation.isError ? (
        <Typography.Text variant="note" color="secondary">
          {describeApiError(joinMutation.error).description}
        </Typography.Text>
      ) : null}
      <Button
        className={s.action}
        size="medium"
        stretched
        tone={joined ? 'secondary' : 'primary'}
        disabled={joined || joinMutation.isPending}
        aria-pressed={joined}
        iconBefore={joined ? <IconCheckCircle /> : <IconPlus />}
        onClick={() => joinMutation.mutate(incidentId)}
      >
        {joined ? incidentTexts.focus.joinedAction : incidentTexts.focus.joinAction}
      </Button>
    </div>
  )
}
