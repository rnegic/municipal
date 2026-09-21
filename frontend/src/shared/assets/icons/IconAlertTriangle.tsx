import type { IconProps } from '../types'

export const IconAlertTriangle = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="currentColor"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M12 2.6 1.4 21.4h21.2L12 2.6Zm1 14.3h-2v-2h2v2Zm0-3.6h-2V9.1h2v4.2Z" />
  </svg>
)
