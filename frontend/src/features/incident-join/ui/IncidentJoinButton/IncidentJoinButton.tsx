import { useState } from 'react'

import { Avatar, Typography } from '@maxhub/max-ui'

import { incidentTexts } from '@/entities/incident'
import { IconCheckCircle, IconPlus } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import s from './IncidentJoinButton.module.scss'

export interface IncidentJoinButtonProps {
  affectedCount: number
  alreadyJoined?: boolean
  className?: string
}

export const IncidentJoinButton = ({
  affectedCount,
  alreadyJoined = false,
  className,
}: IncidentJoinButtonProps) => {
  const [joined, setJoined] = useState(alreadyJoined)
  const count = affectedCount + (joined && !alreadyJoined ? 1 : 0)

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
          {incidentTexts.focus.affectedCount(count)}
        </Typography.Text>
      </div>
      <Button
        className={s.action}
        size="medium"
        stretched
        tone={joined ? 'secondary' : 'primary'}
        disabled={joined}
        aria-pressed={joined}
        iconBefore={joined ? <IconCheckCircle /> : <IconPlus />}
        onClick={() => setJoined(true)}
      >
        {joined ? incidentTexts.focus.joinedAction : incidentTexts.focus.joinAction}
      </Button>
    </div>
  )
}
