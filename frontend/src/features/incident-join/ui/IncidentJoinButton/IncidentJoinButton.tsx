import { Avatar, Typography } from '@maxhub/max-ui'

import { incidentTexts, useJoinIncidentMutation } from '@/entities/incident'
import { IconCheckCircle, IconPlus } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import s from './IncidentJoinButton.module.scss'

export interface IncidentJoinButtonProps {
  incidentId: string
  affectedCount: number
  alreadyJoined?: boolean
  className?: string
}

export const IncidentJoinButton = ({
  incidentId,
  affectedCount,
  alreadyJoined = false,
  className,
}: IncidentJoinButtonProps) => {
  const joinMutation = useJoinIncidentMutation()
  const joined = alreadyJoined || joinMutation.isSuccess

  return (
    <div className={cn(s.root, className)}>
      <div className={s.confirmations}>
        <div className={s.avatars} aria-hidden="true">
          <Avatar.Container size={32}>
            <Avatar.Text gradient="blue">АМ</Avatar.Text>
          </Avatar.Container>
          <Avatar.Container size={32}>
            <Avatar.Text gradient="green">ИК</Avatar.Text>
          </Avatar.Container>
          <Avatar.Container size={32}>
            <Avatar.Text gradient="purple">ОР</Avatar.Text>
          </Avatar.Container>
        </div>
        <Typography.Text variant="note" color="secondary">
          {incidentTexts.focus.affectedCount(affectedCount)}
        </Typography.Text>
      </div>
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
