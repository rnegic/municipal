import { useId, type ChangeEvent, type ReactNode } from 'react'

import { Typography } from '@maxhub/max-ui'

import { commonTexts } from '@/shared/config/texts'
import { cn } from '@/shared/lib/cn'
import s from './Select.module.scss'

export interface SelectOption<TValue extends string = string> {
  value: TValue
  label: string
  description?: string
}

export interface SelectProps<TValue extends string = string> {
  label: string
  options: ReadonlyArray<SelectOption<TValue>>
  value: TValue | ''
  onValueChange: (value: TValue) => void
  placeholder?: string
  hint?: ReactNode
  invalid?: boolean
  disabled?: boolean
  name?: string
  className?: string
}

export const Select = <TValue extends string = string>({
  label,
  options,
  value,
  onValueChange,
  placeholder = commonTexts.select.placeholder,
  hint,
  invalid = false,
  disabled = false,
  name,
  className,
}: SelectProps<TValue>) => {
  const controlId = useId()
  const hintId = useId()

  const handleChange = (event: ChangeEvent<HTMLSelectElement>) => {
    onValueChange(event.target.value as TValue)
  }

  return (
    <div className={cn(s.root, className)}>
      <label className={s.label} htmlFor={controlId}>
        <Typography.Text variant="description" color="secondary">
          {label}
        </Typography.Text>
      </label>
      <select
        id={controlId}
        name={name}
        className={cn(s.control, invalid && s.invalid)}
        value={value}
        disabled={disabled}
        aria-invalid={invalid}
        aria-describedby={hint ? hintId : undefined}
        onChange={handleChange}
      >
        <option value="" disabled>
          {placeholder}
        </option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.description ? `${option.label} — ${option.description}` : option.label}
          </option>
        ))}
      </select>
      {hint ? (
        <span id={hintId} className={cn(s.hint, invalid && s.hintInvalid)}>
          <Typography.Text variant="note" color="inherit">
            {hint}
          </Typography.Text>
        </span>
      ) : null}
    </div>
  )
}
