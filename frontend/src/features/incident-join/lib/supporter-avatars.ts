import { INCIDENT_SUPPORTERS_PREVIEW_COUNT, type IncidentSupporter } from '@/entities/incident'
import { getMaxUserAvatar } from '@/entities/user'
import type { AvatarStackItem } from '@/shared/ui/avatar-stack'

interface SupporterAvatarsParams {
  incidentId: string
  affectedCount: number
  supporters: readonly IncidentSupporter[]
  joined: boolean
  meId?: string | null
}

export const buildSupporterAvatars = ({
  incidentId,
  affectedCount,
  supporters,
  joined,
  meId,
}: SupporterAvatarsParams): AvatarStackItem[] => {
  const me = joined ? getMaxUserAvatar() : null
  const known: AvatarStackItem[] = supporters
    .filter((supporter) => supporter.id !== meId)
    .map((supporter) => ({ id: supporter.id, name: supporter.name, imageUrl: supporter.avatarUrl }))

  const items = me ? [{ id: me.id, name: me.name, imageUrl: me.avatarUrl }, ...known] : known
  const anonymous = Math.min(affectedCount, INCIDENT_SUPPORTERS_PREVIEW_COUNT) - items.length

  return [
    ...items,
    ...Array.from({ length: Math.max(anonymous, 0) }, (_, index) => ({
      id: `${incidentId}-anonymous-${index}`,
    })),
  ].slice(0, Math.max(affectedCount, 1))
}
