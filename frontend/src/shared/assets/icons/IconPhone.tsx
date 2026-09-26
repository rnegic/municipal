import type { IconProps } from '../types'

export const IconPhone = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="currentColor"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M6.54 5c.38 0 .72.24.85.6l1.35 3.75a.9.9 0 0 1-.28 1L6.9 11.6a12.6 12.6 0 0 0 5.5 5.5l1.25-1.56a.9.9 0 0 1 1-.28l3.75 1.35a.9.9 0 0 1 .6.85v2.79a1.6 1.6 0 0 1-1.74 1.6C9.9 21.2 2.8 14.1 2.15 6.74A1.6 1.6 0 0 1 3.75 5h2.79Z" />
  </svg>
)
