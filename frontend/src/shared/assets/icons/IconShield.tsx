import type { IconProps } from '../types'

export const IconShield = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="currentColor"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M12 2 4 5.2v6.1c0 4.9 3.4 9.5 8 10.7 4.6-1.2 8-5.8 8-10.7V5.2L12 2Zm-1.1 14-3.6-3.6 1.5-1.5 2.1 2.1 5-5 1.5 1.5-6.5 6.5Z" />
  </svg>
)
