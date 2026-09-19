import type { IconProps } from './types'

export const IconPencil = ({ size = 20, ...rest }: IconProps) => (
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
    <path d="M4 20h4L19 9a2.83 2.83 0 0 0-4-4L4 16v4Z" />
    <path d="M14.5 5.5 18.5 9.5" />
  </svg>
)
