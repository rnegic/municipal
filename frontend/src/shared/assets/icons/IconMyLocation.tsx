import type { IconProps } from '../types'

export const IconMyLocation = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="none"
    stroke="currentColor"
    strokeWidth={1.8}
    strokeLinecap="round"
    strokeLinejoin="round"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <circle cx="12" cy="12" r="7" />
    <circle cx="12" cy="12" r="2.4" />
    <path d="M12 2v3M12 19v3M2 12h3M19 12h3" />
  </svg>
)
