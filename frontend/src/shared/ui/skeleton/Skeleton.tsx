import { cssVars } from '@/shared/lib/css'
import { cn } from '@/shared/lib/cn'
import s from './Skeleton.module.scss'

export type SkeletonRadius = 'paper' | 'card' | 'round'

export interface SkeletonProps {
  width?: number | string
  height?: number | string
  radius?: SkeletonRadius
  className?: string
}

const RADIUS_CLASS: Record<SkeletonRadius, string> = {
  paper: s.radiusPaper,
  card: s.radiusCard,
  round: s.radiusRound,
}

const toSize = (value: number | string | undefined): string | undefined =>
  typeof value === 'number' ? `${value}px` : value

export const Skeleton = ({ width, height, radius = 'paper', className }: SkeletonProps) => (
  <span
    className={cn(s.root, RADIUS_CLASS[radius], className)}
    style={cssVars({
      '--app-skeleton-width': toSize(width),
      '--app-skeleton-height': toSize(height),
    })}
  />
)
