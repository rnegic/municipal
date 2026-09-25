import { Avatar } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import { getGradient, getInitials } from './helpers'
import s from './AvatarStack.module.scss'

export interface AvatarStackItem {
  id: string
  name?: string | null
  imageUrl?: string | null
}

export interface AvatarStackProps {
  items: readonly AvatarStackItem[]
  total?: number
  max?: number
  size?: number
  label?: string
  className?: string
}

const DEFAULT_MAX = 3
const DEFAULT_SIZE = 32

const renderContent = (item: AvatarStackItem) => {
  const initials = item.name ? getInitials(item.name) : ''

  if (item.imageUrl) {
    return (
      <Avatar.Image
        src={item.imageUrl}
        alt={item.name ?? ''}
        loading="lazy"
        decoding="async"
        fallback={initials || undefined}
        fallbackGradient={getGradient(item.id)}
      />
    )
  }

  if (initials) {
    return <Avatar.Text gradient={getGradient(item.id)}>{initials}</Avatar.Text>
  }

  return <Avatar.Icon />
}

export const AvatarStack = ({
  items,
  total,
  max = DEFAULT_MAX,
  size = DEFAULT_SIZE,
  label,
  className,
}: AvatarStackProps) => {
  if (items.length === 0) {
    return null
  }

  const visible = items.slice(0, max)
  const rest = (total ?? items.length) - visible.length

  return (
    <div className={cn(s.root, className)} role="img" aria-label={label}>
      {visible.map((item) => (
        <Avatar.Container key={item.id} className={s.item} size={size}>
          {renderContent(item)}
        </Avatar.Container>
      ))}
      {rest > 0 ? (
        <Avatar.Container className={s.item} size={size}>
          <Avatar.Text className={s.rest}>+{rest}</Avatar.Text>
        </Avatar.Container>
      ) : null}
    </div>
  )
}
