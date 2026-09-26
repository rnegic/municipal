import type { IconProps } from '../types'

export const IconQrCode = ({ size = 20, ...rest }: IconProps) => (
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
    <rect x="3" y="3" width="6" height="6" rx="1.5" />
    <rect x="15" y="3" width="6" height="6" rx="1.5" />
    <rect x="3" y="15" width="6" height="6" rx="1.5" />
    <path d="M12 3v.01M12 7v3a2 2 0 0 1-2 2M12 12v.01M16 12h1M21 12v.01M12 16v.01M12 21v-1M21 16v-1h-3a2 2 0 0 0-2 2v3M21 21v.01" />
  </svg>
)
