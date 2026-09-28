import type { IconProps } from '../types'

export const IconKey = ({ size = 20, ...rest }: IconProps) => (
  <svg
    viewBox="0 0 24 24"
    width={size}
    height={size}
    fill="currentColor"
    aria-hidden="true"
    focusable="false"
    {...rest}
  >
    <path d="M7.5 15.5a3.5 3.5 0 1 1 3.3-4.7h9.7v3h-2v2.5h-3v-2.5h-4.7a3.5 3.5 0 0 1-3.3 1.7Zm0-2a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z" />
  </svg>
)
