import type { AvatarTextGradient } from '@maxhub/max-ui'

const GRADIENTS: readonly AvatarTextGradient[] = ['blue', 'green', 'purple', 'orange', 'red']

const INITIALS_MAX_LENGTH = 2

/**
 * Builds up to two uppercase initials from a display name.
 *
 * @param name - Display name, possibly with several words.
 * @returns Initials, or an empty string when the name has no letters.
 */
export const getInitials = (name: string): string =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, INITIALS_MAX_LENGTH)
    .map((word) => word[0].toUpperCase())
    .join('')

/**
 * Picks a stable gradient for an avatar so the same person keeps the same colour.
 *
 * @param seed - Stable identifier of the person.
 * @returns One of the MAX UI avatar gradients.
 */
export const getGradient = (seed: string): AvatarTextGradient => {
  const hash = Array.from(seed).reduce((acc, char) => (acc * 31 + char.charCodeAt(0)) % 997, 7)

  return GRADIENTS[hash % GRADIENTS.length]
}
