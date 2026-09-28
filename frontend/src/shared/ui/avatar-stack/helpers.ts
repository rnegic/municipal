import type { AvatarTextGradient } from '@maxhub/max-ui'

const GRADIENTS: readonly AvatarTextGradient[] = ['blue', 'green', 'purple', 'orange', 'red']

const INITIALS_MAX_LENGTH = 2

export const getInitials = (name: string): string =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, INITIALS_MAX_LENGTH)
    .map((word) => word[0].toUpperCase())
    .join('')

export const getGradient = (seed: string): AvatarTextGradient => {
  const hash = Array.from(seed).reduce((acc, char) => (acc * 31 + char.charCodeAt(0)) % 997, 7)

  return GRADIENTS[hash % GRADIENTS.length]
}
