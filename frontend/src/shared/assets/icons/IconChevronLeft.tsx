import type { IconProps } from '../types'

export const IconChevronLeft = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M15 5 8 12l7 7" />
  </svg>
)
