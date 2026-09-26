import type { ComponentProps } from 'react'

import { cn } from '@/shared/lib/cn'
import s from './Checkbox.module.scss'

export interface CheckboxProps extends Omit<ComponentProps<'input'>, 'type'> {
  className?: string
}

export const Checkbox = ({ className, ...rest }: CheckboxProps) => (
  <span className={cn(s.root, className)}>
    <input className={s.input} type="checkbox" {...rest} />
    <span className={s.box} aria-hidden="true" />
  </span>
)
