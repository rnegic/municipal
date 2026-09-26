import { getUser } from '@/shared/lib/max'

export interface MaxUserAvatar {
  id: string
  name: string | null
  avatarUrl: string | null
}

export const getMaxUserAvatar = (): MaxUserAvatar | null => {
  const user = getUser()

  if (!user) {
    return null
  }

  const name = [user.first_name, user.last_name].filter(Boolean).join(' ').trim()

  return {
    id: `max_${user.id}`,
    name: name || user.username || null,
    avatarUrl: user.photo_url ?? null,
  }
}
