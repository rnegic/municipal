import { incidentTexts, type Incident } from '@/entities/incident'
import { IconShare } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import {
  buildMiniAppLink,
  buildShareDeepLink,
  isInsideMax,
  openMaxLink,
  shareMaxContent,
} from '@/shared/lib/max'
import { Button } from '@/shared/ui/button'
import { incidentShareTexts as texts } from '../../config/texts'
import s from './IncidentShareButton.module.scss'

export interface IncidentShareButtonProps {
  incident: Incident
  className?: string
}

export const IncidentShareButton = ({ incident, className }: IncidentShareButtonProps) => {
  if (!isInsideMax()) {
    return null
  }

  const link = buildMiniAppLink(incident.id)
  const text = texts.message({
    severity: incidentTexts.severity[incident.severity],
    title: incident.title,
    status: incidentTexts.statuses[incident.status],
    description: incident.description,
  })

  const handleShare = () => {
    if (shareMaxContent({ text, link: link ?? undefined })) {
      return
    }

    openMaxLink(buildShareDeepLink(link ? `${text}\n${link}` : text))
  }

  return (
    <Button
      className={cn(s.root, className)}
      size="small"
      tone="secondary"
      stretched
      iconBefore={<IconShare size={16} />}
      innerClassNames={{ iconBefore: s.icon, content: s.content }}
      onClick={handleShare}
    >
      {texts.action}
    </Button>
  )
}
