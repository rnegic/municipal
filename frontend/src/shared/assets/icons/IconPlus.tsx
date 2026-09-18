import type { IconProps } from './types'

export const IconPlus = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth={2}
    strokeLinecap="round"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M12 5v14M5 12h14" />
  </svg>
)
