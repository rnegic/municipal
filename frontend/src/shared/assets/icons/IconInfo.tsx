import type { IconProps } from '../types'

export const IconInfo = ({ size = 20, ...rest }: IconProps) => (
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
    <circle cx="12" cy="12" r="10" />
    <path d="M12 11v5" />
    <path d="M12 7.75v.01" />
  </svg>
)
